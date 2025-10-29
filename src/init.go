package main

func init() {
	// Assign the real shutdown function from sbcwaste.go to the placeholder.
	shutdownChromedp = shutdownSbcwasteChromedp
}
