package icepacker

// ListSettings records settings of the listing
type ListSettings struct {
	PackFileName string
	Cipher       CipherSettings
	OnFinish     chan ListResult
}

// Finish returns a success ListResult instance and put to the OnFinish channel if it's not nil
func (this *ListSettings) Finish(err error, fat *FAT) ListResult {
	ret := ListResult{err, fat}
	if this.OnFinish != nil {
		this.OnFinish <- ret
	}
	return ret
}

// Finish returns an errored ListResult instance and put to the OnFinish channel if it's not nil
func (this *ListSettings) FinishError(err error) ListResult {
	ret := ListResult{Err: err}
	if this.OnFinish != nil {
		this.OnFinish <- ret
	}
	return ret
}

// ListPack lists the FAT from the package. Returns a ListResult instance with the FAT
func ListPack(settings ListSettings) ListResult {

	// Hash the cipher key
	shaKey := HashingKey(settings.Cipher)

	// Open the bundle file
	bundle, err := OpenBundle(settings.PackFileName, shaKey)
	if err != nil {
		return settings.FinishError(err)
	}
	defer bundle.Close()

	// List finished
	return settings.Finish(nil, &bundle.FAT)
}
