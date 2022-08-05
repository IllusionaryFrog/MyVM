mod mem;
mod vm;

use gumdrop::Options;
use std::{
    fs::read,
    io::{stdin, Read},
    path::PathBuf,
};

#[derive(Debug, Options)]
struct Args {
    #[options(help = "prints this help message")]
    help: bool,

    #[options(help = "uses file instead of stdin")]
    file: Option<PathBuf>,

    #[options(help = "argument passed to vm program")]
    vargs: Option<String>,
}

fn main() {
    let args = Args::parse_args_default_or_exit();

    let bytes_res = if let Some(input) = args.file {
        read(input)
    } else {
        let mut buf = vec![];
        stdin().lock().read_to_end(&mut buf).map(|_| buf)
    };

    if let Ok(mut bytes) = bytes_res {
        let ptr = bytes.len() as u64;
        if let Some(vargs) = args.vargs {
            let vbytes = vargs.as_bytes();
            bytes.extend(vbytes);
            bytes.extend((vbytes.len() as u64).to_le_bytes());
        } else {
            bytes.extend(0u64.to_le_bytes());
        }
        bytes.extend(ptr.to_le_bytes());

        #[cfg(feature = "debug")]
        let vm = vm::VM::<true>::new(bytes);
        #[cfg(not(feature = "debug"))]
        let vm = vm::VM::<false>::new(bytes);
        vm.run()
    } else {
        eprintln!("Error: Unable to read input");
    }
}
