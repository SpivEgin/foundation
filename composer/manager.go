package composer

// GetComposer returns instance of composer
func GetComposer() InterfaceComposer {
	return registeredComposer
}
