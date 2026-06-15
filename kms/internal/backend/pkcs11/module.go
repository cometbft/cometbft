package pkcs11

import (
	"fmt"
	"sync"

	"github.com/miekg/pkcs11"
)

// A PKCS#11 module is a shared library with process-global C_Initialize state:
// it must be initialized exactly once per process and finalized once. Multiple
// signers may target the same module (e.g. several chains on one HSM), so module
// contexts are ref-counted here and shared across signers. The context is
// finalized only when the last signer using it closes.
var (
	modulesMu sync.Mutex
	modules   = map[string]*moduleRef{}
)

type moduleRef struct {
	ctx  *pkcs11.Ctx
	refs int
}

// acquireModule returns a shared, initialized context for the module at path,
// loading and initializing it on first use and bumping its ref count otherwise.
func acquireModule(path string) (*pkcs11.Ctx, error) {
	modulesMu.Lock()
	defer modulesMu.Unlock()

	if m := modules[path]; m != nil {
		m.refs++
		return m.ctx, nil
	}

	ctx := pkcs11.New(path)
	if ctx == nil {
		return nil, fmt.Errorf("pkcs11: failed to load module %q", path)
	}
	if err := ctx.Initialize(); err != nil {
		// Another consumer in this process may have already initialized the
		// underlying library; treat that as success and adopt it.
		if ce, ok := err.(pkcs11.Error); !ok || ce != pkcs11.CKR_CRYPTOKI_ALREADY_INITIALIZED {
			ctx.Destroy()
			return nil, fmt.Errorf("pkcs11: initialize module %q: %w", path, err)
		}
	}
	modules[path] = &moduleRef{ctx: ctx, refs: 1}
	return ctx, nil
}

// releaseModule drops one reference to the module at path, finalizing and
// unloading it when the last reference is released.
func releaseModule(path string) {
	modulesMu.Lock()
	defer modulesMu.Unlock()

	m := modules[path]
	if m == nil {
		return
	}
	m.refs--
	if m.refs <= 0 {
		_ = m.ctx.Finalize()
		m.ctx.Destroy()
		delete(modules, path)
	}
}
