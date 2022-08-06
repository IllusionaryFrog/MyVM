use crate::mem::MEM;
use std::{
    fmt, fs,
    io::{self, Read, Write},
    thread,
    time::Duration,
};

#[repr(i8)]
pub enum Inter {
    OutOfMem = 001,
    RegOverflow = 002,
    InvalidInst = 003,
    IoError = 004,
}

pub type Result<T> = std::result::Result<T, Inter>;

#[repr(transparent)]
#[derive(Clone, Copy)]
struct Reg(u64);

impl Reg {
    #[inline]
    fn inc_by(&mut self, val: u64) -> Result<()> {
        Ok(*self = self.add(val)?)
    }

    #[inline]
    fn dec_by(&mut self, val: u64) -> Result<()> {
        Ok(*self = self.sub(val)?)
    }

    fn add(self, val: u64) -> Result<Self> {
        let (val, err) = self.0.overflowing_add(val);
        if err {
            Err(Inter::RegOverflow)
        } else {
            Ok(Self(val))
        }
    }

    fn sub(self, val: u64) -> Result<Self> {
        let (val, err) = self.0.overflowing_sub(val);
        if err {
            Err(Inter::RegOverflow)
        } else {
            Ok(Self(val))
        }
    }

    #[inline]
    fn to_bytes(self) -> [u8; 8] {
        self.0.to_le_bytes()
    }

    #[inline]
    fn set(&mut self, val: u64) {
        self.0 = val;
    }

    #[inline]
    fn as_u64(self) -> u64 {
        self.0
    }

    #[inline]
    fn as_usize(self) -> usize {
        self.0 as usize
    }
}

pub struct VM<const DEBUG: bool> {
    pc: Reg,
    sp: Reg,
    cs: Reg,
    ih: Reg,
    ir: i8,
    mem: MEM,
}

impl<const DEBUG: bool> VM<DEBUG> {
    pub fn new(code: Vec<u8>) -> Self {
        Self {
            pc: Reg(0),
            sp: Reg(code.len() as u64 - 16),
            cs: Reg(0),
            ih: Reg(0),
            ir: 0,
            mem: MEM::new(code),
        }
    }

    pub fn run(mut self) {
        loop {
            if let Err(inter) = self._run() {
                self.ir = inter as i8;
                self.pc.set(self.ih.as_u64());
                if DEBUG {
                    println!("inter '{}'", inter as i8);
                    println!("continue...");
                    io::stdin()
                        .read_line(&mut String::new())
                        .expect("continue");
                }
            } else {
                break;
            }
        }
    }

    fn _run(&mut self) -> Result<()> {
        loop {
            let bytes = self.mem.read(self.pc.as_usize())?;
            if DEBUG {
                println!("continue...");
                io::stdin().read_line(&mut String::new()).expect("continue");
                println!("inst '{}'", u8::from_le_bytes(bytes));
                println!("{:?}", self);
            }
            match u8::from_le_bytes(bytes) {
                000 => self.pc.inc_by(1)?,
                001 => break Ok(()),
                002 => {
                    let addr = u64::from_le_bytes(self.pop()?);
                    self.cs.dec_by(8)?;
                    let pc = self.pc.add(1)?.to_bytes();
                    self.mem.write(self.cs.as_usize(), pc)?;
                    self.pc.set(addr);
                }
                003 => {
                    let bytes = self.mem.read(self.cs.as_usize())?;
                    self.cs.inc_by(8)?;
                    self.pc.set(u64::from_le_bytes(bytes));
                }
                004 => self.pc = self.ih,
                005 => {
                    let cap = u64::from_le_bytes(self.pop()?);
                    let ptr = self.mem.alloc(cap as usize) as u64;
                    self.push(ptr.to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                006 => {
                    let len = u64::from_le_bytes(self.pop()?) as usize;
                    let ptr = u64::from_le_bytes(self.pop()?) as usize;
                    let buf = self.mem.slice_mut(ptr, len)?;
                    let read = io::stdin()
                        .lock()
                        .read(buf)
                        .map_err(|_| Inter::IoError)?;
                    self.push((read as u64).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                007 => {
                    let len = u64::from_le_bytes(self.pop()?) as usize;
                    let ptr = u64::from_le_bytes(self.pop()?) as usize;
                    let buf = self.mem.read_bytes(ptr, len)?;
                    let read = io::stdout()
                        .lock()
                        .write(buf)
                        .map_err(|_| Inter::IoError)?;
                    self.push((read as u64).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                008 => {
                    let dest_len = u64::from_le_bytes(self.pop()?) as usize;
                    let dest_ptr = u64::from_le_bytes(self.pop()?) as usize;
                    let path_len = u64::from_le_bytes(self.pop()?) as usize;
                    let path_ptr = u64::from_le_bytes(self.pop()?) as usize;
                    let bytes = self.mem.read_bytes(path_ptr, path_len)?;
                    let path = String::from_utf8_lossy(bytes);
                    let bytes = fs::read(path.as_ref()).unwrap_or_default();
                    let len = dest_len.min(bytes.len());
                    self.mem.write_bytes(dest_ptr, &bytes[..len])?;
                    self.push((len as u64).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                009 => {
                    let src_len = u64::from_le_bytes(self.pop()?) as usize;
                    let src_ptr = u64::from_le_bytes(self.pop()?) as usize;
                    let path_len = u64::from_le_bytes(self.pop()?) as usize;
                    let path_ptr = u64::from_le_bytes(self.pop()?) as usize;
                    let bytes = self.mem.read_bytes(path_ptr, path_len)?;
                    let path = String::from_utf8_lossy(bytes);
                    let bytes = self.mem.read_bytes(src_ptr, src_len)?;
                    let len = fs::write(path.as_ref(), bytes)
                        .map(|_| src_len)
                        .unwrap_or_default();
                    self.push((len as u64).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }

                010 => {
                    self.pc.inc_by(1)?;
                    let val = self.mem.read(self.pc.as_usize())?;
                    self.push::<1>(val)?;
                    self.pc.inc_by(1)?;
                }
                011 => {
                    self.pc.inc_by(1)?;
                    let val = self.mem.read(self.pc.as_usize())?;
                    self.push::<2>(val)?;
                    self.pc.inc_by(2)?;
                }
                012 => {
                    self.pc.inc_by(1)?;
                    let val = self.mem.read(self.pc.as_usize())?;
                    self.push::<4>(val)?;
                    self.pc.inc_by(4)?;
                }
                013 => {
                    self.pc.inc_by(1)?;
                    let val = self.mem.read(self.pc.as_usize())?;
                    self.push::<8>(val)?;
                    self.pc.inc_by(8)?;
                }
                014 => {
                    self.pc.inc_by(1)?;
                    let val = self.mem.read(self.pc.as_usize())?;
                    self.push::<16>(val)?;
                    self.pc.inc_by(16)?;
                }
                015 => {
                    let val = u64::from_le_bytes(self.pop()?);
                    self.sp.set(val);
                    self.pc.inc_by(1)?;
                }
                016 => {
                    let val = u64::from_le_bytes(self.pop()?);
                    self.cs.set(val);
                    self.pc.inc_by(1)?;
                }
                017 => {
                    let val = u64::from_le_bytes(self.pop()?);
                    self.ih.set(val);
                    self.pc.inc_by(1)?;
                }
                018 => {
                    self.ir = i8::from_le_bytes(self.pop()?);
                    self.pc.inc_by(1)?;
                }
                019 => {
                    self.push(self.ir.to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }

                020 => {
                    let _ = self.pop::<1>()?;
                    self.pc.inc_by(1)?;
                }
                021 => {
                    let _ = self.pop::<2>()?;
                    self.pc.inc_by(1)?;
                }
                022 => {
                    let _ = self.pop::<4>()?;
                    self.pc.inc_by(1)?;
                }
                023 => {
                    let _ = self.pop::<8>()?;
                    self.pc.inc_by(1)?;
                }
                024 => {
                    let _ = self.pop::<16>()?;
                    self.pc.inc_by(1)?;
                }
                025 => {
                    let v = i8::from_le_bytes(self.pop()?);
                    self.push(v.saturating_neg().to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                026 => {
                    let v = i16::from_le_bytes(self.pop()?);
                    self.push(v.saturating_neg().to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                027 => {
                    let v = i32::from_le_bytes(self.pop()?);
                    self.push(v.saturating_neg().to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                028 => {
                    let v = i64::from_le_bytes(self.pop()?);
                    self.push(v.saturating_neg().to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                029 => {
                    let v = i128::from_le_bytes(self.pop()?);
                    self.push(v.saturating_neg().to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }

                030 => {
                    let v1 = self.pop()?;
                    let v2 = self.pop()?;
                    self.push::<1>(v1)?;
                    self.push::<1>(v2)?;
                    self.pc.inc_by(1)?;
                }
                031 => {
                    let v1 = self.pop()?;
                    let v2 = self.pop()?;
                    self.push::<2>(v1)?;
                    self.push::<2>(v2)?;
                    self.pc.inc_by(1)?;
                }
                032 => {
                    let v1 = self.pop()?;
                    let v2 = self.pop()?;
                    self.push::<4>(v1)?;
                    self.push::<4>(v2)?;
                    self.pc.inc_by(1)?;
                }
                033 => {
                    let v1 = self.pop()?;
                    let v2 = self.pop()?;
                    self.push::<8>(v1)?;
                    self.push::<8>(v2)?;
                    self.pc.inc_by(1)?;
                }
                034 => {
                    let v1 = self.pop()?;
                    let v2 = self.pop()?;
                    self.push::<16>(v1)?;
                    self.push::<16>(v2)?;
                    self.pc.inc_by(1)?;
                }
                035 => {
                    let v1 = self.pop()?;
                    let v2 = self.pop()?;
                    let v3 = self.pop()?;
                    self.push::<1>(v2)?;
                    self.push::<1>(v1)?;
                    self.push::<1>(v3)?;
                    self.pc.inc_by(1)?;
                }
                036 => {
                    let v1 = self.pop()?;
                    let v2 = self.pop()?;
                    let v3 = self.pop()?;
                    self.push::<2>(v2)?;
                    self.push::<2>(v1)?;
                    self.push::<2>(v3)?;
                    self.pc.inc_by(1)?;
                }
                037 => {
                    let v1 = self.pop()?;
                    let v2 = self.pop()?;
                    let v3 = self.pop()?;
                    self.push::<4>(v2)?;
                    self.push::<4>(v1)?;
                    self.push::<4>(v3)?;
                    self.pc.inc_by(1)?;
                }
                038 => {
                    let v1 = self.pop()?;
                    let v2 = self.pop()?;
                    let v3 = self.pop()?;
                    self.push::<8>(v2)?;
                    self.push::<8>(v1)?;
                    self.push::<8>(v3)?;
                    self.pc.inc_by(1)?;
                }
                039 => {
                    let v1 = self.pop()?;
                    let v2 = self.pop()?;
                    let v3 = self.pop()?;
                    self.push::<16>(v2)?;
                    self.push::<16>(v1)?;
                    self.push::<16>(v3)?;
                    self.pc.inc_by(1)?;
                }

                040 => {
                    let val = self.mem.read(self.sp.as_usize())?;
                    self.push::<1>(val)?;
                    self.pc.inc_by(1)?;
                }
                041 => {
                    let val = self.mem.read(self.sp.as_usize())?;
                    self.push::<2>(val)?;
                    self.pc.inc_by(1)?;
                }
                042 => {
                    let val = self.mem.read(self.sp.as_usize())?;
                    self.push::<4>(val)?;
                    self.pc.inc_by(1)?;
                }
                043 => {
                    let val = self.mem.read(self.sp.as_usize())?;
                    self.push::<8>(val)?;
                    self.pc.inc_by(1)?;
                }
                044 => {
                    let val = self.mem.read(self.sp.as_usize())?;
                    self.push::<16>(val)?;
                    self.pc.inc_by(1)?;
                }
                045 => {
                    let sp = self.sp.add(1)?.as_usize();
                    self.push::<1>(self.mem.read(sp)?)?;
                    self.pc.inc_by(1)?;
                }
                046 => {
                    let sp = self.sp.add(2)?.as_usize();
                    self.push::<2>(self.mem.read(sp)?)?;
                    self.pc.inc_by(1)?;
                }
                047 => {
                    let sp = self.sp.add(4)?.as_usize();
                    self.push::<4>(self.mem.read(sp)?)?;
                    self.pc.inc_by(1)?;
                }
                048 => {
                    let sp = self.sp.add(8)?.as_usize();
                    self.push::<8>(self.mem.read(sp)?)?;
                    self.pc.inc_by(1)?;
                }
                049 => {
                    let sp = self.sp.add(16)?.as_usize();
                    self.push::<16>(self.mem.read(sp)?)?;
                    self.pc.inc_by(1)?;
                }

                050 => {
                    let v2 = u8::from_le_bytes(self.pop()?);
                    let v1 = u8::from_le_bytes(self.pop()?);
                    self.push((v1 & v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                051 => {
                    let v2 = u16::from_le_bytes(self.pop()?);
                    let v1 = u16::from_le_bytes(self.pop()?);
                    self.push((v1 & v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                052 => {
                    let v2 = u32::from_le_bytes(self.pop()?);
                    let v1 = u32::from_le_bytes(self.pop()?);
                    self.push((v1 & v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                053 => {
                    let v2 = u64::from_le_bytes(self.pop()?);
                    let v1 = u64::from_le_bytes(self.pop()?);
                    self.push((v1 & v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                054 => {
                    let v2 = u128::from_le_bytes(self.pop()?);
                    let v1 = u128::from_le_bytes(self.pop()?);
                    self.push((v1 & v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                055 => {
                    let v2 = u8::from_le_bytes(self.pop()?);
                    let v1 = u8::from_le_bytes(self.pop()?);
                    self.push((v1 | v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                056 => {
                    let v2 = u16::from_le_bytes(self.pop()?);
                    let v1 = u16::from_le_bytes(self.pop()?);
                    self.push((v1 | v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                057 => {
                    let v2 = u32::from_le_bytes(self.pop()?);
                    let v1 = u32::from_le_bytes(self.pop()?);
                    self.push((v1 | v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                058 => {
                    let v2 = u64::from_le_bytes(self.pop()?);
                    let v1 = u64::from_le_bytes(self.pop()?);
                    self.push((v1 | v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                059 => {
                    let v2 = u128::from_le_bytes(self.pop()?);
                    let v1 = u128::from_le_bytes(self.pop()?);
                    self.push((v1 | v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }

                060 => {
                    let v2 = u8::from_le_bytes(self.pop()?);
                    let v1 = u8::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_shl(v2 as u32).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                061 => {
                    let v2 = u8::from_le_bytes(self.pop()?);
                    let v1 = u16::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_shl(v2 as u32).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                062 => {
                    let v2 = u8::from_le_bytes(self.pop()?);
                    let v1 = u32::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_shl(v2 as u32).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                063 => {
                    let v2 = u8::from_le_bytes(self.pop()?);
                    let v1 = u64::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_shl(v2 as u32).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                064 => {
                    let v2 = u8::from_le_bytes(self.pop()?);
                    let v1 = u128::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_shl(v2 as u32).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                065 => {
                    let v2 = u8::from_le_bytes(self.pop()?);
                    let v1 = u8::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_shr(v2 as u32).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                066 => {
                    let v2 = u8::from_le_bytes(self.pop()?);
                    let v1 = u16::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_shr(v2 as u32).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                067 => {
                    let v2 = u8::from_le_bytes(self.pop()?);
                    let v1 = u32::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_shr(v2 as u32).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                068 => {
                    let v2 = u8::from_le_bytes(self.pop()?);
                    let v1 = u64::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_shr(v2 as u32).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                069 => {
                    let v2 = u8::from_le_bytes(self.pop()?);
                    let v1 = u128::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_shr(v2 as u32).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }

                070 => {
                    let v2 = u8::from_le_bytes(self.pop()?);
                    let v1 = u8::from_le_bytes(self.pop()?);
                    self.push(v1.rotate_left(v2 as u32).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                071 => {
                    let v2 = u8::from_le_bytes(self.pop()?);
                    let v1 = u16::from_le_bytes(self.pop()?);
                    self.push(v1.rotate_left(v2 as u32).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                072 => {
                    let v2 = u8::from_le_bytes(self.pop()?);
                    let v1 = u32::from_le_bytes(self.pop()?);
                    self.push(v1.rotate_left(v2 as u32).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                073 => {
                    let v2 = u8::from_le_bytes(self.pop()?);
                    let v1 = u64::from_le_bytes(self.pop()?);
                    self.push(v1.rotate_left(v2 as u32).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                074 => {
                    let v2 = u8::from_le_bytes(self.pop()?);
                    let v1 = u128::from_le_bytes(self.pop()?);
                    self.push(v1.rotate_left(v2 as u32).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                075 => {
                    let v2 = u8::from_le_bytes(self.pop()?);
                    let v1 = u8::from_le_bytes(self.pop()?);
                    self.push(v1.rotate_right(v2 as u32).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                076 => {
                    let v2 = u8::from_le_bytes(self.pop()?);
                    let v1 = u16::from_le_bytes(self.pop()?);
                    self.push(v1.rotate_right(v2 as u32).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                077 => {
                    let v2 = u8::from_le_bytes(self.pop()?);
                    let v1 = u32::from_le_bytes(self.pop()?);
                    self.push(v1.rotate_right(v2 as u32).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                078 => {
                    let v2 = u8::from_le_bytes(self.pop()?);
                    let v1 = u64::from_le_bytes(self.pop()?);
                    self.push(v1.rotate_right(v2 as u32).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                079 => {
                    let v2 = u8::from_le_bytes(self.pop()?);
                    let v1 = u128::from_le_bytes(self.pop()?);
                    self.push(v1.rotate_right(v2 as u32).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }

                080 => {
                    let v2 = u8::from_le_bytes(self.pop()?);
                    let v1 = u8::from_le_bytes(self.pop()?);
                    self.push(((v1 == v2) as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                081 => {
                    let v2 = u16::from_le_bytes(self.pop()?);
                    let v1 = u16::from_le_bytes(self.pop()?);
                    self.push(((v1 == v2) as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                082 => {
                    let v2 = u32::from_le_bytes(self.pop()?);
                    let v1 = u32::from_le_bytes(self.pop()?);
                    self.push(((v1 == v2) as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                083 => {
                    let v2 = u64::from_le_bytes(self.pop()?);
                    let v1 = u64::from_le_bytes(self.pop()?);
                    self.push(((v1 == v2) as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                084 => {
                    let v2 = u128::from_le_bytes(self.pop()?);
                    let v1 = u128::from_le_bytes(self.pop()?);
                    self.push(((v1 == v2) as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                085 => {
                    let v2 = u8::from_le_bytes(self.pop()?);
                    let v1 = u8::from_le_bytes(self.pop()?);
                    self.push(((v1 != v2) as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                086 => {
                    let v2 = u16::from_le_bytes(self.pop()?);
                    let v1 = u16::from_le_bytes(self.pop()?);
                    self.push(((v1 != v2) as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                087 => {
                    let v2 = u32::from_le_bytes(self.pop()?);
                    let v1 = u32::from_le_bytes(self.pop()?);
                    self.push(((v1 != v2) as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                088 => {
                    let v2 = u64::from_le_bytes(self.pop()?);
                    let v1 = u64::from_le_bytes(self.pop()?);
                    self.push(((v1 != v2) as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                089 => {
                    let v2 = u128::from_le_bytes(self.pop()?);
                    let v1 = u128::from_le_bytes(self.pop()?);
                    self.push(((v1 != v2) as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }

                090 => {
                    let addr = u64::from_le_bytes(self.pop()?);
                    self.pc.set(addr);
                }
                091 => {
                    let off = u64::from_le_bytes(self.pop()?);
                    self.pc.inc_by(off)?;
                }
                092 => {
                    let off = u64::from_le_bytes(self.pop()?);
                    self.pc.dec_by(off)?;
                }
                093 => break Err(Inter::InvalidInst),
                094 => {
                    let milis = u64::from_le_bytes(self.pop()?);
                    thread::sleep(Duration::from_millis(milis));
                    self.pc.inc_by(1)?;
                }
                095 => {
                    let con = u8::from_le_bytes(self.pop()?);
                    let addr = u64::from_le_bytes(self.pop()?);
                    if con != 0 {
                        self.pc.set(addr);
                    } else {
                        self.pc.inc_by(1)?;
                    }
                }
                096 => {
                    let con = u8::from_le_bytes(self.pop()?);
                    let off = u64::from_le_bytes(self.pop()?);
                    if con != 0 {
                        self.pc.inc_by(off)?;
                    } else {
                        self.pc.inc_by(1)?;
                    }
                }
                097 => {
                    let con = u8::from_le_bytes(self.pop()?);
                    let off = u64::from_le_bytes(self.pop()?);
                    if con != 0 {
                        self.pc.dec_by(off)?;
                    } else {
                        self.pc.inc_by(1)?;
                    }
                }
                098 | 099 => break Err(Inter::InvalidInst),

                100 => {
                    let v2 = u8::from_le_bytes(self.pop()?);
                    let v1 = u8::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_add(v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                101 => {
                    let v2 = u16::from_le_bytes(self.pop()?);
                    let v1 = u16::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_add(v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                102 => {
                    let v2 = u32::from_le_bytes(self.pop()?);
                    let v1 = u32::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_add(v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                103 => {
                    let v2 = u64::from_le_bytes(self.pop()?);
                    let v1 = u64::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_add(v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                104 => {
                    let v2 = u128::from_le_bytes(self.pop()?);
                    let v1 = u128::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_add(v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                105 => {
                    let v2 = i8::from_le_bytes(self.pop()?);
                    let v1 = i8::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_add(v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                106 => {
                    let v2 = i16::from_le_bytes(self.pop()?);
                    let v1 = i16::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_add(v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                107 => {
                    let v2 = i32::from_le_bytes(self.pop()?);
                    let v1 = i32::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_add(v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                108 => {
                    let v2 = i64::from_le_bytes(self.pop()?);
                    let v1 = i64::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_add(v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                109 => {
                    let v2 = i128::from_le_bytes(self.pop()?);
                    let v1 = i128::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_add(v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }

                110 => {
                    let v2 = u8::from_le_bytes(self.pop()?);
                    let v1 = u8::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_sub(v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                111 => {
                    let v2 = u16::from_le_bytes(self.pop()?);
                    let v1 = u16::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_sub(v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                112 => {
                    let v2 = u32::from_le_bytes(self.pop()?);
                    let v1 = u32::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_sub(v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                113 => {
                    let v2 = u64::from_le_bytes(self.pop()?);
                    let v1 = u64::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_sub(v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                114 => {
                    let v2 = u128::from_le_bytes(self.pop()?);
                    let v1 = u128::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_sub(v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                115 => {
                    let v2 = i8::from_le_bytes(self.pop()?);
                    let v1 = i8::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_sub(v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                116 => {
                    let v2 = i16::from_le_bytes(self.pop()?);
                    let v1 = i16::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_sub(v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                117 => {
                    let v2 = i32::from_le_bytes(self.pop()?);
                    let v1 = i32::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_sub(v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                118 => {
                    let v2 = i64::from_le_bytes(self.pop()?);
                    let v1 = i64::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_sub(v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                119 => {
                    let v2 = i128::from_le_bytes(self.pop()?);
                    let v1 = i128::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_sub(v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }

                120 => {
                    let v2 = u8::from_le_bytes(self.pop()?);
                    let v1 = u8::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_mul(v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                121 => {
                    let v2 = u16::from_le_bytes(self.pop()?);
                    let v1 = u16::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_mul(v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                122 => {
                    let v2 = u32::from_le_bytes(self.pop()?);
                    let v1 = u32::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_mul(v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                123 => {
                    let v2 = u64::from_le_bytes(self.pop()?);
                    let v1 = u64::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_mul(v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                124 => {
                    let v2 = u128::from_le_bytes(self.pop()?);
                    let v1 = u128::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_mul(v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                125 => {
                    let v2 = i8::from_le_bytes(self.pop()?);
                    let v1 = i8::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_mul(v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                126 => {
                    let v2 = i16::from_le_bytes(self.pop()?);
                    let v1 = i16::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_mul(v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                127 => {
                    let v2 = i32::from_le_bytes(self.pop()?);
                    let v1 = i32::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_mul(v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                128 => {
                    let v2 = i64::from_le_bytes(self.pop()?);
                    let v1 = i64::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_mul(v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                129 => {
                    let v2 = i128::from_le_bytes(self.pop()?);
                    let v1 = i128::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_mul(v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }

                130 => {
                    let v2 = u8::from_le_bytes(self.pop()?);
                    let v1 = u8::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_div(v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                131 => {
                    let v2 = u16::from_le_bytes(self.pop()?);
                    let v1 = u16::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_div(v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                132 => {
                    let v2 = u32::from_le_bytes(self.pop()?);
                    let v1 = u32::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_div(v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                133 => {
                    let v2 = u64::from_le_bytes(self.pop()?);
                    let v1 = u64::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_div(v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                134 => {
                    let v2 = u128::from_le_bytes(self.pop()?);
                    let v1 = u128::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_div(v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                135 => {
                    let v2 = i8::from_le_bytes(self.pop()?);
                    let v1 = i8::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_div(v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                136 => {
                    let v2 = i16::from_le_bytes(self.pop()?);
                    let v1 = i16::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_div(v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                137 => {
                    let v2 = i32::from_le_bytes(self.pop()?);
                    let v1 = i32::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_div(v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                138 => {
                    let v2 = i64::from_le_bytes(self.pop()?);
                    let v1 = i64::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_div(v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                139 => {
                    let v2 = i128::from_le_bytes(self.pop()?);
                    let v1 = i128::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_div(v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }

                140 => {
                    let v2 = u8::from_le_bytes(self.pop()?);
                    let v1 = u8::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_rem(v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                141 => {
                    let v2 = u16::from_le_bytes(self.pop()?);
                    let v1 = u16::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_rem(v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                142 => {
                    let v2 = u32::from_le_bytes(self.pop()?);
                    let v1 = u32::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_rem(v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                143 => {
                    let v2 = u64::from_le_bytes(self.pop()?);
                    let v1 = u64::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_rem(v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                144 => {
                    let v2 = u128::from_le_bytes(self.pop()?);
                    let v1 = u128::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_rem(v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                145 => {
                    let v2 = i8::from_le_bytes(self.pop()?);
                    let v1 = i8::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_rem(v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                146 => {
                    let v2 = i16::from_le_bytes(self.pop()?);
                    let v1 = i16::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_rem(v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                147 => {
                    let v2 = i32::from_le_bytes(self.pop()?);
                    let v1 = i32::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_rem(v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                148 => {
                    let v2 = i64::from_le_bytes(self.pop()?);
                    let v1 = i64::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_rem(v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                149 => {
                    let v2 = i128::from_le_bytes(self.pop()?);
                    let v1 = i128::from_le_bytes(self.pop()?);
                    self.push(v1.wrapping_rem(v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }

                150 => {
                    let v2 = u8::from_le_bytes(self.pop()?);
                    let v1 = u8::from_le_bytes(self.pop()?);
                    self.push(((v1 < v2) as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                151 => {
                    let v2 = u16::from_le_bytes(self.pop()?);
                    let v1 = u16::from_le_bytes(self.pop()?);
                    self.push(((v1 < v2) as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                152 => {
                    let v2 = u32::from_le_bytes(self.pop()?);
                    let v1 = u32::from_le_bytes(self.pop()?);
                    self.push(((v1 < v2) as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                153 => {
                    let v2 = u64::from_le_bytes(self.pop()?);
                    let v1 = u64::from_le_bytes(self.pop()?);
                    self.push(((v1 < v2) as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                154 => {
                    let v2 = u128::from_le_bytes(self.pop()?);
                    let v1 = u128::from_le_bytes(self.pop()?);
                    self.push(((v1 < v2) as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                155 => {
                    let v2 = i8::from_le_bytes(self.pop()?);
                    let v1 = i8::from_le_bytes(self.pop()?);
                    self.push(((v1 < v2) as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                156 => {
                    let v2 = i16::from_le_bytes(self.pop()?);
                    let v1 = i16::from_le_bytes(self.pop()?);
                    self.push(((v1 < v2) as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                157 => {
                    let v2 = i32::from_le_bytes(self.pop()?);
                    let v1 = i32::from_le_bytes(self.pop()?);
                    self.push(((v1 < v2) as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                158 => {
                    let v2 = i64::from_le_bytes(self.pop()?);
                    let v1 = i64::from_le_bytes(self.pop()?);
                    self.push(((v1 < v2) as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                159 => {
                    let v2 = i128::from_le_bytes(self.pop()?);
                    let v1 = i128::from_le_bytes(self.pop()?);
                    self.push(((v1 < v2) as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }

                160 => {
                    let v2 = u8::from_le_bytes(self.pop()?);
                    let v1 = u8::from_le_bytes(self.pop()?);
                    self.push(((v1 <= v2) as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                161 => {
                    let v2 = u16::from_le_bytes(self.pop()?);
                    let v1 = u16::from_le_bytes(self.pop()?);
                    self.push(((v1 <= v2) as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                162 => {
                    let v2 = u32::from_le_bytes(self.pop()?);
                    let v1 = u32::from_le_bytes(self.pop()?);
                    self.push(((v1 <= v2) as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                163 => {
                    let v2 = u64::from_le_bytes(self.pop()?);
                    let v1 = u64::from_le_bytes(self.pop()?);
                    self.push(((v1 <= v2) as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                164 => {
                    let v2 = u128::from_le_bytes(self.pop()?);
                    let v1 = u128::from_le_bytes(self.pop()?);
                    self.push(((v1 <= v2) as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                165 => {
                    let v2 = i8::from_le_bytes(self.pop()?);
                    let v1 = i8::from_le_bytes(self.pop()?);
                    self.push(((v1 <= v2) as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                166 => {
                    let v2 = i16::from_le_bytes(self.pop()?);
                    let v1 = i16::from_le_bytes(self.pop()?);
                    self.push(((v1 <= v2) as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                167 => {
                    let v2 = i32::from_le_bytes(self.pop()?);
                    let v1 = i32::from_le_bytes(self.pop()?);
                    self.push(((v1 <= v2) as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                168 => {
                    let v2 = i64::from_le_bytes(self.pop()?);
                    let v1 = i64::from_le_bytes(self.pop()?);
                    self.push(((v1 <= v2) as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                169 => {
                    let v2 = i128::from_le_bytes(self.pop()?);
                    let v1 = i128::from_le_bytes(self.pop()?);
                    self.push(((v1 <= v2) as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }

                170 => {
                    let v2 = u8::from_le_bytes(self.pop()?);
                    let v1 = u8::from_le_bytes(self.pop()?);
                    self.push(((v1 > v2) as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                171 => {
                    let v2 = u16::from_le_bytes(self.pop()?);
                    let v1 = u16::from_le_bytes(self.pop()?);
                    self.push(((v1 > v2) as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                172 => {
                    let v2 = u32::from_le_bytes(self.pop()?);
                    let v1 = u32::from_le_bytes(self.pop()?);
                    self.push(((v1 > v2) as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                173 => {
                    let v2 = u64::from_le_bytes(self.pop()?);
                    let v1 = u64::from_le_bytes(self.pop()?);
                    self.push(((v1 > v2) as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                174 => {
                    let v2 = u128::from_le_bytes(self.pop()?);
                    let v1 = u128::from_le_bytes(self.pop()?);
                    self.push(((v1 > v2) as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                175 => {
                    let v2 = i8::from_le_bytes(self.pop()?);
                    let v1 = i8::from_le_bytes(self.pop()?);
                    self.push(((v1 > v2) as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                176 => {
                    let v2 = i16::from_le_bytes(self.pop()?);
                    let v1 = i16::from_le_bytes(self.pop()?);
                    self.push(((v1 > v2) as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                177 => {
                    let v2 = i32::from_le_bytes(self.pop()?);
                    let v1 = i32::from_le_bytes(self.pop()?);
                    self.push(((v1 > v2) as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                178 => {
                    let v2 = i64::from_le_bytes(self.pop()?);
                    let v1 = i64::from_le_bytes(self.pop()?);
                    self.push(((v1 > v2) as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                179 => {
                    let v2 = i128::from_le_bytes(self.pop()?);
                    let v1 = i128::from_le_bytes(self.pop()?);
                    self.push(((v1 > v2) as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }

                180 => {
                    let v2 = u8::from_le_bytes(self.pop()?);
                    let v1 = u8::from_le_bytes(self.pop()?);
                    self.push(((v1 >= v2) as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                181 => {
                    let v2 = u16::from_le_bytes(self.pop()?);
                    let v1 = u16::from_le_bytes(self.pop()?);
                    self.push(((v1 >= v2) as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                182 => {
                    let v2 = u32::from_le_bytes(self.pop()?);
                    let v1 = u32::from_le_bytes(self.pop()?);
                    self.push(((v1 >= v2) as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                183 => {
                    let v2 = u64::from_le_bytes(self.pop()?);
                    let v1 = u64::from_le_bytes(self.pop()?);
                    self.push(((v1 >= v2) as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                184 => {
                    let v2 = u128::from_le_bytes(self.pop()?);
                    let v1 = u128::from_le_bytes(self.pop()?);
                    self.push(((v1 >= v2) as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                185 => {
                    let v2 = i8::from_le_bytes(self.pop()?);
                    let v1 = i8::from_le_bytes(self.pop()?);
                    self.push(((v1 >= v2) as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                186 => {
                    let v2 = i16::from_le_bytes(self.pop()?);
                    let v1 = i16::from_le_bytes(self.pop()?);
                    self.push(((v1 >= v2) as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                187 => {
                    let v2 = i32::from_le_bytes(self.pop()?);
                    let v1 = i32::from_le_bytes(self.pop()?);
                    self.push(((v1 >= v2) as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                188 => {
                    let v2 = i64::from_le_bytes(self.pop()?);
                    let v1 = i64::from_le_bytes(self.pop()?);
                    self.push(((v1 >= v2) as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                189 => {
                    let v2 = i128::from_le_bytes(self.pop()?);
                    let v1 = i128::from_le_bytes(self.pop()?);
                    self.push(((v1 >= v2) as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }

                190 => {
                    let v = u8::from_le_bytes(self.pop()?);
                    self.push((v as u16).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                191 => {
                    let v = u8::from_le_bytes(self.pop()?);
                    self.push((v as u32).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                192 => {
                    let v = u8::from_le_bytes(self.pop()?);
                    self.push((v as u64).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                193 => {
                    let v = u8::from_le_bytes(self.pop()?);
                    self.push((v as u128).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                194 => {
                    let v = u16::from_le_bytes(self.pop()?);
                    self.push((v as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                195 => {
                    let v = u16::from_le_bytes(self.pop()?);
                    self.push((v as u32).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                196 => {
                    let v = u16::from_le_bytes(self.pop()?);
                    self.push((v as u64).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                197 => {
                    let v = u16::from_le_bytes(self.pop()?);
                    self.push((v as u128).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                198 => {
                    let v = u32::from_le_bytes(self.pop()?);
                    self.push((v as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                199 => {
                    let v = u32::from_le_bytes(self.pop()?);
                    self.push((v as u16).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }

                200 => {
                    let v = u32::from_le_bytes(self.pop()?);
                    self.push((v as u64).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                201 => {
                    let v = u32::from_le_bytes(self.pop()?);
                    self.push((v as u128).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                202 => {
                    let v = u64::from_le_bytes(self.pop()?);
                    self.push((v as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                203 => {
                    let v = u64::from_le_bytes(self.pop()?);
                    self.push((v as u16).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                204 => {
                    let v = u64::from_le_bytes(self.pop()?);
                    self.push((v as u32).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                205 => {
                    let v = u64::from_le_bytes(self.pop()?);
                    self.push((v as u128).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                206 => {
                    let v = u128::from_le_bytes(self.pop()?);
                    self.push((v as u8).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                207 => {
                    let v = u128::from_le_bytes(self.pop()?);
                    self.push((v as u16).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                208 => {
                    let v = u128::from_le_bytes(self.pop()?);
                    self.push((v as u32).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                209 => {
                    let v = u128::from_le_bytes(self.pop()?);
                    self.push((v as u64).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }

                210 => {
                    let ptr = u64::from_le_bytes(self.pop()?) as usize;
                    let val = self.mem.read(ptr)?;
                    self.push::<1>(val)?;
                    self.pc.inc_by(1)?;
                }
                211 => {
                    let ptr = u64::from_le_bytes(self.pop()?) as usize;
                    let val = self.mem.read(ptr)?;
                    self.push::<2>(val)?;
                    self.pc.inc_by(1)?;
                }
                212 => {
                    let ptr = u64::from_le_bytes(self.pop()?) as usize;
                    let val = self.mem.read(ptr)?;
                    self.push::<4>(val)?;
                    self.pc.inc_by(1)?;
                }
                213 => {
                    let ptr = u64::from_le_bytes(self.pop()?) as usize;
                    let val = self.mem.read(ptr)?;
                    self.push::<8>(val)?;
                    self.pc.inc_by(1)?;
                }
                214 => {
                    let ptr = u64::from_le_bytes(self.pop()?) as usize;
                    let val = self.mem.read(ptr)?;
                    self.push::<16>(val)?;
                    self.pc.inc_by(1)?;
                }
                215 => {
                    let val = self.pop::<1>()?;
                    let ptr = u64::from_le_bytes(self.pop()?) as usize;
                    self.mem.write(ptr, val)?;
                    self.pc.inc_by(1)?;
                }
                216 => {
                    let val = self.pop::<2>()?;
                    let ptr = u64::from_le_bytes(self.pop()?) as usize;
                    self.mem.write(ptr, val)?;
                    self.pc.inc_by(1)?;
                }
                217 => {
                    let val = self.pop::<4>()?;
                    let ptr = u64::from_le_bytes(self.pop()?) as usize;
                    self.mem.write(ptr, val)?;
                    self.pc.inc_by(1)?;
                }
                218 => {
                    let val = self.pop::<8>()?;
                    let ptr = u64::from_le_bytes(self.pop()?) as usize;
                    self.mem.write(ptr, val)?;
                    self.pc.inc_by(1)?;
                }
                219 => {
                    let val = self.pop::<16>()?;
                    let ptr = u64::from_le_bytes(self.pop()?) as usize;
                    self.mem.write(ptr, val)?;
                    self.pc.inc_by(1)?;
                }

                220 => {
                    let pc = self.pc.add(1)?;
                    let bytes = self.mem.read(pc.as_usize())?;
                    self.pc.set(u64::from_le_bytes(bytes));
                }
                221 => {
                    let pc = self.pc.add(1)?;
                    let bytes = self.mem.read(pc.as_usize())?;
                    self.pc.inc_by(u64::from_le_bytes(bytes))?;
                }
                222 => {
                    let pc = self.pc.add(1)?;
                    let bytes = self.mem.read(pc.as_usize())?;
                    self.pc.dec_by(u64::from_le_bytes(bytes))?;
                }
                223 => break Err(Inter::InvalidInst),
                224 => {
                    let pc = self.pc.add(1)?;
                    let bytes = self.mem.read(pc.as_usize())?;
                    let milis = u64::from_le_bytes(bytes);
                    thread::sleep(Duration::from_millis(milis));
                    self.pc.inc_by(9)?;
                }
                225 => {
                    let con = u8::from_le_bytes(self.pop()?);
                    if con != 0 {
                        let pc = self.pc.add(1)?;
                        let bytes = self.mem.read(pc.as_usize())?;
                        self.pc.set(u64::from_le_bytes(bytes));
                    } else {
                        self.pc.inc_by(9)?;
                    }
                }
                226 => {
                    let con = u8::from_le_bytes(self.pop()?);
                    if con != 0 {
                        let pc = self.pc.add(1)?;
                        let bytes = self.mem.read(pc.as_usize())?;
                        self.pc.inc_by(u64::from_le_bytes(bytes))?;
                    } else {
                        self.pc.inc_by(9)?;
                    }
                }
                227 => {
                    let con = u8::from_le_bytes(self.pop()?);
                    if con != 0 {
                        let pc = self.pc.add(1)?;
                        let bytes = self.mem.read(pc.as_usize())?;
                        self.pc.dec_by(u64::from_le_bytes(bytes))?;
                    } else {
                        self.pc.inc_by(9)?;
                    }
                }
                228 => break Err(Inter::InvalidInst),
                229 => {
                    let pc = self.pc.add(1)?;
                    let bytes = self.mem.read(pc.as_usize())?;
                    self.cs.dec_by(8)?;
                    let pc = self.pc.add(9)?.to_bytes();
                    self.mem.write(self.cs.as_usize(), pc)?;
                    self.pc.set(u64::from_le_bytes(bytes));
                }

                230 => {
                    let pc = self.pc.add(1)?;
                    let bytes = self.mem.read(pc.as_usize())?;
                    let ptr = u64::from_le_bytes(bytes) as usize;
                    let val = self.mem.read(ptr)?;
                    self.push::<1>(val)?;
                    self.pc.inc_by(9)?;
                }
                231 => {
                    let pc = self.pc.add(1)?;
                    let bytes = self.mem.read(pc.as_usize())?;
                    let ptr = u64::from_le_bytes(bytes) as usize;
                    let val = self.mem.read(ptr)?;
                    self.push::<2>(val)?;
                    self.pc.inc_by(9)?;
                }
                232 => {
                    let pc = self.pc.add(1)?;
                    let bytes = self.mem.read(pc.as_usize())?;
                    let ptr = u64::from_le_bytes(bytes) as usize;
                    let val = self.mem.read(ptr)?;
                    self.push::<4>(val)?;
                    self.pc.inc_by(9)?;
                }
                233 => {
                    let pc = self.pc.add(1)?;
                    let bytes = self.mem.read(pc.as_usize())?;
                    let ptr = u64::from_le_bytes(bytes) as usize;
                    let val = self.mem.read(ptr)?;
                    self.push::<8>(val)?;
                    self.pc.inc_by(9)?;
                }
                234 => {
                    let pc = self.pc.add(1)?;
                    let bytes = self.mem.read(pc.as_usize())?;
                    let ptr = u64::from_le_bytes(bytes) as usize;
                    let val = self.mem.read(ptr)?;
                    self.push::<16>(val)?;
                    self.pc.inc_by(9)?;
                }
                235 => {
                    let val = self.pop::<1>()?;
                    let pc = self.pc.add(1)?;
                    let bytes = self.mem.read(pc.as_usize())?;
                    let ptr = u64::from_le_bytes(bytes) as usize;
                    self.mem.write(ptr, val)?;
                    self.pc.inc_by(9)?;
                }
                236 => {
                    let val = self.pop::<2>()?;
                    let pc = self.pc.add(1)?;
                    let bytes = self.mem.read(pc.as_usize())?;
                    let ptr = u64::from_le_bytes(bytes) as usize;
                    self.mem.write(ptr, val)?;
                    self.pc.inc_by(9)?;
                }
                237 => {
                    let val = self.pop::<4>()?;
                    let pc = self.pc.add(1)?;
                    let bytes = self.mem.read(pc.as_usize())?;
                    let ptr = u64::from_le_bytes(bytes) as usize;
                    self.mem.write(ptr, val)?;
                    self.pc.inc_by(9)?;
                }
                238 => {
                    let val = self.pop::<8>()?;
                    let pc = self.pc.add(1)?;
                    let bytes = self.mem.read(pc.as_usize())?;
                    let ptr = u64::from_le_bytes(bytes) as usize;
                    self.mem.write(ptr, val)?;
                    self.pc.inc_by(9)?;
                }
                239 => {
                    let val = self.pop::<16>()?;
                    let pc = self.pc.add(1)?;
                    let bytes = self.mem.read(pc.as_usize())?;
                    let ptr = u64::from_le_bytes(bytes) as usize;
                    self.mem.write(ptr, val)?;
                    self.pc.inc_by(9)?;
                }

                240 => {
                    let v2 = u8::from_le_bytes(self.pop()?);
                    let v1 = u8::from_le_bytes(self.pop()?);
                    self.push((v1 ^ v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                241 => {
                    let v2 = u16::from_le_bytes(self.pop()?);
                    let v1 = u16::from_le_bytes(self.pop()?);
                    self.push((v1 ^ v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                242 => {
                    let v2 = u32::from_le_bytes(self.pop()?);
                    let v1 = u32::from_le_bytes(self.pop()?);
                    self.push((v1 ^ v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                243 => {
                    let v2 = u64::from_le_bytes(self.pop()?);
                    let v1 = u64::from_le_bytes(self.pop()?);
                    self.push((v1 ^ v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }
                244 => {
                    let v2 = u128::from_le_bytes(self.pop()?);
                    let v1 = u128::from_le_bytes(self.pop()?);
                    self.push((v1 ^ v2).to_le_bytes())?;
                    self.pc.inc_by(1)?;
                }

                245..=249 => break Err(Inter::InvalidInst),

                250 => {
                    println!("{:?}", self);
                    self.pc.inc_by(1)?;
                }
                251 => {
                    let val = u8::from_le_bytes(self.pop()?);
                    println!("Debug8: 0x{:x}", val);
                    self.pc.inc_by(1)?;
                }
                252 => {
                    let val = u16::from_le_bytes(self.pop()?);
                    println!("Debug16: 0x{:x}", val);
                    self.pc.inc_by(1)?;
                }
                253 => {
                    let val = u32::from_le_bytes(self.pop()?);
                    println!("Debug32: 0x{:x}", val);
                    self.pc.inc_by(1)?;
                }
                254 => {
                    let val = u64::from_le_bytes(self.pop()?);
                    println!("Debug64: 0x{:x}", val);
                    self.pc.inc_by(1)?;
                }
                255 => {
                    let val = u128::from_le_bytes(self.pop()?);
                    println!("Debug128: 0x{:x}", val);
                    self.pc.inc_by(1)?;
                }
            }
        }
    }

    fn push<const N: usize>(&mut self, val: [u8; N]) -> Result<()> {
        self.sp.dec_by(N as u64)?;
        self.mem.write(self.sp.as_usize(), val)
    }

    fn pop<const N: usize>(&mut self) -> Result<[u8; N]> {
        let res = self.mem.read(self.sp.as_usize())?;
        self.sp.inc_by(N as u64)?;
        Ok(res)
    }
}

impl<const DEBUG: bool> fmt::Debug for VM<DEBUG> {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        writeln!(
            f,
            "VM {{ pc: 0x{:x}, sp: 0x{:x}, cs: 0x{:x}, ih: 0x{:x}, ir: {} }}",
            self.pc.as_u64(),
            self.sp.as_u64(),
            self.cs.as_u64(),
            self.ih.as_u64(),
            self.ir,
        )?;
        self.mem.fmt(f)
    }
}
