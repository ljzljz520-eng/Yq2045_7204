# BUG_REPRO

The following failures were observed while validating the initial project state.
Each section records what failed, how to reproduce it, and the complete command output.
They are preserved intentionally; only failing build gates are omitted from the generated Dockerfile.

## Failure 1: Go test (.)

- Observed problem: `Go test (.)` failed in the initial project state.
- Working directory: `.`
- Command: `cd /app && GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go test -count=1 ./...`
- Exit status: `1`

```text
ok  	sitepay	0.268s
?   	sitepay/cmd/payroll	[no test files]
?   	sitepay/internal/config	[no test files]
ok  	sitepay/internal/domain	0.004s
ok  	sitepay/internal/importer	0.008s
--- FAIL: TestPayrollKeepsCents (0.00s)
    payroll_test.go:19: night allowance cents=1200, want 1275
FAIL
FAIL	sitepay/internal/payroll	0.017s
ok  	sitepay/internal/report	0.012s
ok  	sitepay/internal/service	0.156s
ok  	sitepay/internal/store	0.123s
ok  	sitepay/internal/validation	0.005s
FAIL
```

## Architecture reproduction

### linux/amd64
- Go toolchain version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/payroll): exit `0`
### linux/arm64
- Go toolchain version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/payroll): exit `0`
