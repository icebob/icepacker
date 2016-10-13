package icepacker

// ProgressState records the state of packing or unpacking
type ProgressState struct {
	Err         error
	Total       int
	Index       int
	CurrentFile string
}

// FinishResult records some information about packing or unpacking
type FinishResult struct {
	Err       error
	FileCount int64
	Size      int64
	DupCount  int
	DupSize   int64
}

// ListResult records the result of the listing
type ListResult struct {
	Err error
	FAT *FAT
}
