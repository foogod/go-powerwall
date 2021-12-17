// Package powerwall implements an interface for accessing Tesla Powerwall
// devices via the local-network API.
//
// Note: This API is currently undocumented, and all information about it has
// been obtained through reverse-engineering.  Some information here may
// therefore be incomplete or incorrect at this time, and it is also possible
// that the API will change without warning.
//
// Much of the information used to build this library was obtained from
// https://github.com/vloschiavo/powerwall2.  If you have additional
// information about any of this, please help by contributing what you know to
// that project.
//
// General usage:
//
// First, you will need to create a new powerwall.Client object using the
// NewClient function.  After that has been done, you can use any of the
// following functions on that object to interact with the Powerwall gateway.
package powerwall
