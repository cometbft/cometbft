package types

// Adapter interface for adapter
type Adapter interface {
	Id() string
	Perform(job OracleJob, result AdapterResult, runTimeInput AdapterRunTimeInput, store *AdapterStore) (AdapterResult, error)
	Validate(job OracleJob) error
}

// AdapterResult struct for adapter results
type AdapterResult struct {
	Data map[string]GenericValue
}

// GetData get data for given key
func (result AdapterResult) GetData(key string) GenericValue {
	return result.Data[key]
}

// SetData sets data for the given key, value pair
func (result *AdapterResult) SetData(key string, value GenericValue) {
	result.Data[key] = value
}

// NewAdapterResult returns an initialized AdapterResult
func NewAdapterResult() AdapterResult {
	result := AdapterResult{}
	result.Data = make(map[string]GenericValue)
	return result
}

// AdapterRunTimeInput struct for adapter input
type AdapterRunTimeInput struct {
	LastStoreData       map[string]GenericValue
	LastStoreDataExists bool
	BeginTime           uint64
	Config              CustomNodeConfig
}

// GetLastStoreData gets data for the given key
func (input AdapterRunTimeInput) GetLastStoreData(key string) GenericValue {
	return input.LastStoreData[key]
}

// AdapterStore struct for adapter store
type AdapterStore struct {
	ShouldPersist bool
	Data          map[string]GenericValue
}

// NewAdapterStore returns an initialized adapter store
func NewAdapterStore() AdapterStore {
	store := AdapterStore{}
	store.Data = make(map[string]GenericValue)
	return store
}

// SetData sets data for the given key, value pair
func (store *AdapterStore) SetData(key string, value GenericValue) {
	store.Data[key] = value
}
