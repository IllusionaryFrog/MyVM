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
}

fn main() {
    let args = Args::parse_args_default_or_exit();

    let bytes_res = if let Some(input) = args.file {
        read(input)
    } else {
        let mut buf = vec![];
        stdin().lock().read_to_end(&mut buf).map(|_| buf)
    };

    if let Ok(bytes) = bytes_res {
        #[cfg(feature = "debug")]
        let vm = vm::VM::<true>::new(bytes);
        #[cfg(not(feature = "debug"))]
        let vm = vm::VM::<false>::new(bytes);
        vm.run()
    } else {
        eprintln!("Error: Unable to read input");
    }
}
