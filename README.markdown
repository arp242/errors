Yet another `errors` package for Go.

Import as `zgo.at/errors`; [godoc](https://pkg.go.dev/zgo.at/errors).

This is based on the new errors API introduced in Go 1.13 with the additions:

1. Add `Wrap(err, "some context")` and `Wrapf(err, "some context", ...)`, which
   return nil if the passed error is nil.

   I tried using the `fmt.Errorf("...: %w", err)` pattern, but I found I missed
   being able to do:

       func f() error {
           return errors.Wrap(f2(), "context")
       }

   Which I find much more convenient than:

       func f() error {
           err := f2()
           if err != nil {
               return fmt.Errorf("context: %w", err)
           }
           return nil
       }

2. A stack trace is added with `erorrs.New()`, `errors.Errorf()`, and
   `erorrs.Wrap()` I know it's not needed with appropriate context but sometimes
   I accidentally add the same context more than once, or just want to quickly
   see where *exactly* the error is coming from. Especially on dev it's much
   more convenient.

   You can use `errors.Package` to only add stack traces for a specific package.
   For example `errors.Package = "zgo.at/goatcounter"` will only add trace lines
   in `zgo.at/goatcounter/...`

   You can control the maximum stack size with `errors.StackSize`; set to `0` to
   disable adding stack traces altogether (i.e. on production).

3. `Group` type for collecting grouped errors:

        errs := NewGroup(20)  // 20 is the maximum amount of errors
        for {
            err := doStuff()
            if err.Append(err) { // Won't append on nil, returns false if it's nil.
                continue
            }
        }

        fmt.Println(errs)        // Print all errors (if any).
        return errs.ErrorOrNil() // No errors? Returns nil.
