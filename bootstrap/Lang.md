# Lang

### Spec
- `Ast`     => `(Fun)*`
- `Fun`     => `"fun" Opts? Ident "(" Args ")" Block`
- `Opts`    => `"{" (Ident),* "}"`
- `Args`    => `(Typ),* ":" (Typ),*`
- `Block`   => `"{" (Let)* (Expr)* "}"`
- `Typ`     => `"u8" | ... | "u128" | "i8" | ... | "i128"`
- `Expr`     => `Ident | String | Number`