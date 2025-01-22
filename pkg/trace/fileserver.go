package trace

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const jsonL = ".jsonl"

func (lt *LocalTracer) getTableHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the request to get the data
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		inputString := r.FormValue("table")
		if inputString == "" {
			http.Error(w, "No data provided", http.StatusBadRequest)
			return
		}

		f, done, err := lt.readTable(inputString)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to read table: %v", err), http.StatusInternalServerError)
			return
		}
		defer done() //nolint:errcheck

		// Use the pump function to continuously read from the file and write to
		// the response writer
		reader, writer := pump(inputString, bufio.NewReader(f))
		defer reader.Close()

		// Set the content type to the writer's form data content type
		w.Header().Set("Content-Type", writer.FormDataContentType())

		// Copy the data from the reader to the response writer
		if _, err := io.Copy(w, reader); err != nil {
			http.Error(w, "Failed to send data", http.StatusInternalServerError)
			return
		}
	}
}

// pump continuously reads from a bufio.Reader and writes to a multipart.Writer.
// It returns the reader end of the pipe and the writer for consumption by the
// server.
func pump(table string, br *bufio.Reader) (*io.PipeReader, *multipart.Writer) {
	r, w := io.Pipe()
	m := multipart.NewWriter(w)

	go func(
		table string,
		m *multipart.Writer,
		w *io.PipeWriter,
		br *bufio.Reader,
	) {
		defer w.Close()
		defer m.Close()

		part, err := m.CreateFormFile("filename", table+jsonL)
		if err != nil {
			return
		}

		if _, err = io.Copy(part, br); err != nil {
			return
		}

	}(table, m, w, br)

	return r, m
}

func (lt *LocalTracer) servePullData() {
	mux := http.NewServeMux()
	mux.HandleFunc("/get_table", lt.getTableHandler())
	err := http.ListenAndServe(lt.cfg.Instrumentation.TracePullAddress, mux) //nolint:gosec
	if err != nil {
		lt.logger.Error("trace pull server failure", "err", err)
	}
	lt.logger.Info("trace pull server started", "address", lt.cfg.Instrumentation.TracePullAddress)
}

// GetTable downloads a table from the server and saves it to the given directory. It uses a multipart
// response to download the file.
func GetTable(serverURL, table, dirPath string) error {
	data := url.Values{}
	data.Set("table", table)

	serverURL = serverURL + "/get_table"

	resp, err := http.PostForm(serverURL, data) //nolint:gosec
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	_, params, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil {
		return err
	}

	boundary, ok := params["boundary"]
	if !ok {
		panic("Not a multipart response")
	}

	err = os.MkdirAll(dirPath, 0755)
	if err != nil {
		return err
	}

	outputFile, err := os.Create(path.Join(dirPath, table+jsonL))
	if err != nil {
		return err
	}
	defer outputFile.Close()

	reader := multipart.NewReader(resp.Body, boundary)

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break // End of multipart
		}
		if err != nil {
			return err
		}

		contentDisposition, params, err := mime.ParseMediaType(part.Header.Get("Content-Disposition"))
		if err != nil {
			return err
		}

		if contentDisposition == "form-data" && params["filename"] != "" {
			_, err = io.Copy(outputFile, part)
			if err != nil {
				return err
			}
		}

		part.Close()
	}

	return nil
}

// S3Config is a struct that holds the configuration for an S3 bucket.
type S3Config struct {
	BucketName string `json:"bucket_name"`
	Region     string `json:"region"`
	AccessKey  string `json:"access_key"`
	SecretKey  string `json:"secret_key"`
	// PushDelay is the time in seconds to wait before pushing the file to S3.
	// If this is 0, it defaults is used.
	PushDelay int64 `json:"push_delay"`
}

// readS3Config reads an S3Config from a file in the given directory.
func readS3Config(dir string) (S3Config, error) {
	cfg := S3Config{}
	f, err := os.Open(filepath.Join(dir, "s3.json"))
	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return cfg, err
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(&cfg)
	if cfg.PushDelay == 0 {
		cfg.PushDelay = 60
	}
	return cfg, err
}

// PushS3 pushes a file to an S3 bucket using the given S3Config. It uses the
// chainID and the nodeID to organize the files in the bucket. The directory
// structure is chainID/nodeID/table.jsonl .
func PushS3(chainID, nodeID string, s3cfg S3Config, f *os.File) error {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(s3cfg.Region),
		Credentials: credentials.NewStaticCredentials(
			s3cfg.AccessKey,
			s3cfg.SecretKey,
			"",
		),
		HTTPClient: &http.Client{
			Timeout: time.Duration(15) * time.Second,
		},
	},
	)
	if err != nil {
		return err
	}

	s3Svc := s3.New(sess)

	key := fmt.Sprintf("%s/%s/%s", chainID, nodeID, filepath.Base(f.Name()))

	_, err = s3Svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(s3cfg.BucketName),
		Key:    aws.String(key),
		Body:   f,
	})

	return err
}

func (lt *LocalTracer) pushLoop() {
	for {
		time.Sleep(time.Second * time.Duration(lt.s3Config.PushDelay))
		err := lt.PushAll()
		if err != nil {
			lt.logger.Error("failed to push tables", "error", err)
		}
	}
}

func (lt *LocalTracer) PushAll() error {
	for table := range lt.fileMap {
		f, done, err := lt.readTable(table)
		if err != nil {
			return err
		}
		for i := 0; i < 3; i++ {
			err = PushS3(lt.chainID, lt.nodeID, lt.s3Config, f)
			if err == nil {
				break
			}
			lt.logger.Error("failed to push table", "table", table, "error", err)
			time.Sleep(time.Second * time.Duration(rand.Intn(3))) //nolint:gosec
		}
		err = done()
		if err != nil {
			return err
		}
	}
	return nil
}

// S3Download downloads files that match some prefix from an S3 bucket to a
// local directory dst.
// fileNames is a list of traced jsonl file names to download. If it is empty, all traces are downloaded.
// fileNames should not have .jsonl suffix.
func S3Download(dst, prefix string, cfg S3Config, fileNames ...string) error {
	// Ensure local directory structure exists
	err := os.MkdirAll(dst, os.ModePerm)
	if err != nil {
		return err
	}

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(cfg.Region),
		Credentials: credentials.NewStaticCredentials(
			cfg.AccessKey,
			cfg.SecretKey,
			"",
		),
	},
	)
	if err != nil {
		return err
	}

	s3Svc := s3.New(sess)
	input := &s3.ListObjectsV2Input{
		Bucket:    aws.String(cfg.BucketName),
		Prefix:    aws.String(prefix),
		Delimiter: aws.String(""),
	}

	err = s3Svc.ListObjectsV2Pages(input, func(page *s3.ListObjectsV2Output, lastPage bool) bool {
		for _, content := range page.Contents {
			key := *content.Key

			// If no fileNames are specified, download all files
			if len(fileNames) == 0 {
				fileNames = append(fileNames, strings.TrimPrefix(key, prefix))
			}

			for _, filename := range fileNames {
				// Add .jsonl suffix to the fileNames
				fullFilename := filename + jsonL
				if strings.HasSuffix(key, fullFilename) {
					localFilePath := filepath.Join(dst, prefix, strings.TrimPrefix(key, prefix))
					fmt.Printf("Downloading %s to %s\n", key, localFilePath)

					// Create the directories in the path
					if err := os.MkdirAll(filepath.Dir(localFilePath), os.ModePerm); err != nil {
						return false
					}

					// Create a file to write the S3 Object contents to.
					f, err := os.Create(localFilePath)
					if err != nil {
						return false
					}

					resp, err := s3Svc.GetObject(&s3.GetObjectInput{
						Bucket: aws.String(cfg.BucketName),
						Key:    aws.String(key),
					})
					if err != nil {
						f.Close()
						continue
					}
					defer resp.Body.Close()

					// Copy the contents of the S3 object to the local file
					if _, err := io.Copy(f, resp.Body); err != nil {
						f.Close()
						return false
					}

					fmt.Printf("Successfully downloaded %s to %s\n", key, localFilePath)
					f.Close()
				}
			}
		}
		return !lastPage // continue paging
	})
	return err
}
