echo "Formatting protobuf files"
find ./ -name "*.proto" -exec clang-format -i {} \;

set -e

echo "Generating gogo proto code"
cd proto
proto_dirs=$(find ./Switcheo/carbon -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do
  for file in $(find "${dir}" -maxdepth 1 -name '*.proto'); do
    # Check if the go_package in the file is pointing to carbon
    if grep -q "option go_package.*cometbft" "$file"; then
      buf generate --template buf.gen.gogo.yaml "$file"
    fi
  done
done

cd ..

# move proto files to the right places
cp -r github.com/cometbft/cometbft/* ./
rm -rf github.com
