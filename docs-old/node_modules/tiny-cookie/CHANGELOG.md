### v2.3.2

- **Fix**: Set path to root when testing if cookies are enabled #28

### v2.3.1

- **Fix**: Fix path option is ignored when removeCookie is called #27

### v2.3.0

- **Feature**: Add flow types.

### v2.2.0

- **Feature**: Add typescript types.
- **Document**: Add Chinese document.

### v2.1.2

- **Fix**: Fix `es` directory missing bug #22.

### v2.1.1

- **Fix**: Fix bug when passing options as the third parameter to set method the encoder is null #18.

### v2.1.0

- **Feature**: The `remove()` method supports configuring the domain parameter.

### v2.0.2

- **Fix**: Fix the es modules build [#16](https://github.com/Alex1990/tiny-cookie/issues/16)

### v2.0.1

- **Fix**: Fix the "main" entry in package.json [#13](https://github.com/Alex1990/tiny-cookie/issues/13)

### v2.0.0

With modern development workflow, such as Babel, Rollup, Karma, npm scripts and so on.

- **Breaking change**: Do not support the `Cookie` as a function.
- **Breaking change**: There is not a default export. That is, `import cookie from 'tiny-cookie` doesn't work. The reason why it hasn't a default export is it will prevent the webpack tree-shaking working. You can do it like this `import * as cookie from 'tiny-cookie'`.[#14](https://github.com/Alex1990/tiny-cookie/issues/14)
- **Breaking change**: Rename `enabled` method to `isEnabled`.
- **Feature**: Add `getAll` method to get all cookie pairs at a time.
- **Feature**: Add aliases for all methods, for details, you can see [API](https://github.com/Alex1990/tiny-cookie#apis)
- Add a command to start an http(s) server.
