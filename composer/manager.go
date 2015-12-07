package composer

// GetComposer returns composer
func GetComposer() InterfaceComposer {
	return registeredComposer
}
