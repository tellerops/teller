### v3.2.5

* Add missing CustomData support for entries

### v3.2.4

* Add support for handling protected value unlocking with `Entry` or `Group` being loaded first from XML
* Initialize fresh UUIDs on unmarshal in case they are missing

### v3.2.3

* Adds `(*Binary).GetContentString() (string, error)` and `(*Binary).GetContentBytes() ([]byte, error)` funcs
* Deprecates `(*Binary).GetContent() (string, error)`
* Also adds `CustomIcon` support on `Group` level

### v3.2.2

* Correctly support multiple Window Associations in an entry's AutoType data

### v3.2.1

* Add missing DefaultSequence in AutoType data

### v3.2.0

* Add support for custom icons

### v3.1.0

* Add initialization support for KDBXv4 files
* Add SettingsChanged MetaData field

### v3.0.5

* Improve time marshalling/unmarshalling performance

### v3.0.4

* Ensure time values are formatted according to the version when encoding the DB to file
* Split up code into several smaller files

### v3.0.3

* Split up `BoolWrapper` and `NullableBoolWrapper`

### v3.0.2

* Improve AES decrypt performance (cont.)

### v3.0.1

* Improve AES decrypt performance

### v3.0.0

* Fix `BoolWrapper` to support null values
    - This introduced a breaking change

### v2.1.3

* Fix `TimeWrapper` marshalling and unmarshalling

### v2.1.2

* Attempt to fix `TimeWrapper`

### v2.1.1

* Add `ParseKeyData` to allow loading keys without file operation

### v2.1.0

* Add functional option support for all kinds of initializers

### v2.0.3

* Add KDBX4 HMAC verification on file decoding

### v2.0.2

* Fix KDBX4 HMAC building for encrypted content blocks on file encoding

### v2.0.1

* Drop counter for SalsaStream

### v2.0.0

* KDBX v4.0 support
* Argon2 support
* ChaCha20 support
* Restructured code
* Fixed support for keyfile
* Moved type wrappers into separate package

### v1.0.0

* KDBX v3.1 support
