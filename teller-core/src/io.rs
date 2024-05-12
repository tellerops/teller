use std::{
    io::{self, Read},
    path::Path,
};

use fs_err::File;

pub fn is_binary_file(path: &Path) -> io::Result<bool> {
    let mut file = File::open(path)?;
    let mut buffer = [0; 1024]; // Read the first 1024 bytes of the file

    let bytes_read = file.read(&mut buffer)?;
    for item in buffer.iter().take(bytes_read) {
        if *item == 0 {
            return Ok(true); // Found a null byte, indicating a binary file
        }
    }

    Ok(false) // No null byte found, likely a text file.
}
