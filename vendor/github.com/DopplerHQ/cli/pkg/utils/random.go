/*
Copyright © 2020 Doppler <support@doppler.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package utils

import (
	"crypto/rand"
	"encoding/base64"
	"math"
)

// RandomBase64String cryptographically secure random string
// from https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
func RandomBase64String(l int) string {
	// Base 64 text is 1/3 longer than base 256. (2^8 vs 2^6; 8bits/6bits = 1.333 ratio)
	buffer := make([]byte, int(math.Round(float64(l)/float64(4/3))))
	rand.Read(buffer) // #nosec G104
	str := base64.RawURLEncoding.EncodeToString(buffer)
	return str[:l] // strip 1 extra character we get from odd length results
}
