//go:build darwin && !ios
// +build darwin,!ios

package keychain

/*
#cgo LDFLAGS: -framework CoreFoundation -framework Security
#include <CoreFoundation/CoreFoundation.h>
#include <Security/Security.h>
*/
import "C"

// AccessibleKey is key for kSecAttrAccessible
var AccessibleKey = attrKey(C.CFTypeRef(C.kSecAttrAccessible))
var accessibleTypeRef = map[Accessible]C.CFTypeRef{
	AccessibleWhenUnlocked:                   C.CFTypeRef(C.kSecAttrAccessibleWhenUnlocked),
	AccessibleAfterFirstUnlock:               C.CFTypeRef(C.kSecAttrAccessibleAfterFirstUnlock),
	AccessibleAlways:                         C.CFTypeRef(C.kSecAttrAccessibleAlways),
	AccessibleWhenUnlockedThisDeviceOnly:     C.CFTypeRef(C.kSecAttrAccessibleWhenUnlockedThisDeviceOnly),
	AccessibleAfterFirstUnlockThisDeviceOnly: C.CFTypeRef(C.kSecAttrAccessibleAfterFirstUnlockThisDeviceOnly),
	AccessibleAccessibleAlwaysThisDeviceOnly: C.CFTypeRef(C.kSecAttrAccessibleAlwaysThisDeviceOnly),

	// Only available in 10.10
	//AccessibleWhenPasscodeSetThisDeviceOnly:  C.CFTypeRef(C.kSecAttrAccessibleWhenPasscodeSetThisDeviceOnly),
}

// DeleteItemRef deletes a keychain item reference.
func DeleteItemRef(ref C.CFTypeRef) error {
	errCode := C.SecKeychainItemDelete(C.SecKeychainItemRef(ref))
	return checkError(errCode)
}
