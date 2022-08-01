mod mem;
mod vm;

use gumdrop::Options;
use std::{
    fs::read,
    io::{stdin, Read},
    path::PathBuf,
    process::exit,
};

#[derive(Debug, Options)]
struct Args {
    #[options(help = "prints this help message")]
    help: bool,

    #[options(help = "uses input file to read bytes")]
    input: Option<PathBuf>,

    #[options(help = "uses stdin to read bytes")]
    stdin: bool,
}

fn main() {
    let args = Args::parse_args_default_or_exit();

    let bytes = if args.stdin == args.input.is_some() {
        eprintln!("{}", Args::usage());
        drop(args);
        exit(2);
    } else if let Some(input) = args.input {
        read(input).expect("read")
    } else {
        let mut buf = vec![];
        stdin().lock().read_to_end(&mut buf).expect("read");
        buf
    };

    vm::VM::new(bytes).run()
}
