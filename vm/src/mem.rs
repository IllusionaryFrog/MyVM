use crate::vm::{Inter, Result};
use std::fmt;

pub struct MEM {
    bytes: Vec<u8>,
}

impl MEM {
    pub fn new(bytes: Vec<u8>) -> Self {
        Self { bytes }
    }

    pub fn alloc(&mut self, additional: usize) -> usize {
        let len = self.bytes.len();
        self.bytes.resize(len + additional, 0);
        len
    }

    pub fn read<const N: usize>(&self, addr: usize) -> Result<[u8; N]> {
        if addr + N <= self.bytes.len() {
            Ok(unsafe { *(self.bytes.as_ptr().add(addr) as *const [u8; N]) })
        } else {
            Err(Inter::OutOfMem)
        }
    }

    pub fn write<const N: usize>(
        &mut self,
        addr: usize,
        val: [u8; N],
    ) -> Result<()> {
        if addr + N <= self.bytes.len() {
            Ok(unsafe {
                *(self.bytes.as_mut_ptr().add(addr) as *mut [u8; N]) = val
            })
        } else {
            Err(Inter::OutOfMem)
        }
    }

    pub fn read_bytes(&self, addr: usize, len: usize) -> Result<&[u8]> {
        if addr + len <= self.bytes.len() {
            Ok(unsafe {
                let ptr = &*self.bytes.as_ptr().add(addr);
                std::slice::from_raw_parts(ptr, len)
            })
        } else {
            Err(Inter::OutOfMem)
        }
    }

    pub fn write_bytes(&mut self, addr: usize, bytes: &[u8]) -> Result<()> {
        let len = bytes.len();
        if addr + len <= self.bytes.len() {
            let slice = unsafe {
                let ptr = &mut *self.bytes.as_mut_ptr().add(addr);
                std::slice::from_raw_parts_mut(ptr, len)
            };
            Ok(slice.copy_from_slice(bytes))
        } else {
            Err(Inter::OutOfMem)
        }
    }

    pub fn slice_mut(&mut self, addr: usize, len: usize) -> Result<&mut [u8]> {
        if addr + len <= self.bytes.len() {
            Ok(unsafe {
                let ptr = &mut *self.bytes.as_mut_ptr().add(addr);
                std::slice::from_raw_parts_mut(ptr, len)
            })
        } else {
            Err(Inter::OutOfMem)
        }
    }
}

impl fmt::Debug for MEM {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        let mut rhx = rhexdump::Rhexdump::default();
        rhx.set_format("#[OFFSET]: #[RAW]").unwrap();
        rhx.display_duplicate_lines(false);
        writeln!(f, "{}", rhx.hexdump(&self.bytes))
    }
}
