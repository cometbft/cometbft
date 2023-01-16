const {
  ensureLeadingSlash,
  ensureEndingSlash,
  compose,
  logger,
  chalk,
  path,
  fs,
} = require('@vuepress/shared-utils')

const isInternalUrl = url => url.startsWith('/');

function onlyInternalUrl(fn) {
  return function (url) {
    if (isInternalUrl(url)) {
      return fn(url);
    }
    return url;
  }
}

module.exports = (options = {}, ctx) => {
  const redirectsFile = path.join(ctx.vuepressDir, 'redirects')
  if (!fs.existsSync(redirectsFile)) {
    return
  }

  const { countdown = 3 } = options

  const normalize = compose(
    onlyInternalUrl(ensureEndingSlash),
    onlyInternalUrl(ensureLeadingSlash),
    v => v.trim(),
    v => decodeURIComponent(v),
  )

  const redirects = fs.readFileSync(path.join(ctx.vuepressDir, 'redirects'), 'utf-8')
    .trim()
    .split('\n')
    .map(line => line.split(' ').map(normalize))

  function findRedirect(url) {
    const conf = redirects.find(r => r[0] === url)
    return conf && conf[1]
  }

  const plugin = {}

  async function renderPage(url, redirectUrl) {
    const pagePath = decodeURIComponent(url)
    const filename = pagePath.replace(/\/$/, '/index.html').replace(/^\//, '')
    const filePath = path.resolve(ctx.outDir, filename)
    await fs.ensureDir(path.dirname(filePath))
    const html = getRedirectHtml({ url, redirectUrl, countdown, routerBase: ctx.base })
    await fs.writeFile(filePath, html)
    logger.info(`[redirect] Generated redirect page: ${chalk.cyan(filename)}:`)
    logger.info(`[redirect] ${chalk.gray(url)} -> ${chalk.gray(redirectUrl)}`)
  }

  if (ctx.isProd) {
    plugin.generated = async () => {
      await Promise.all(redirects.map(r => renderPage(r[0], r[1])))
    }
  } else {
    plugin.beforeDevServer = app => {
      app.use((req, res, next) => {
        const redirectUrl = findRedirect(normalize(req.url))
        if (redirectUrl) {
          res.redirect(redirectUrl)
        } else {
          next()
        }
      })
    }
  }

  return plugin
}

/**
 * Get redirect html content.
 */
function getRedirectHtml({ /* url, */redirectUrl, countdown, routerBase = '/' } = {}) {

  return `<!doctype html>
<html lang="en">
<head>
  <meta charset="UTF-8">
</head>
<body>
  ${countdown > 0
      ? `
<p align="center" style="font-size: 30px;">
  Redirect after <b><span id="countdown">${countdown}</span></b>s
</p>
<p align="center">
  <small>Powered by <a target="_blank" href="https://github.com/vuepressjs/vuepress-plugin-html-redirect">vuepress-plugin-redirect</a></small>
</p>`.trim()
      : ''
    }
  <script> 
  var el = document.getElementById("countdown");
  var isIE = navigator.appName.indexOf("Explorer") > -1;
  var routerBase = window.routerBase || ${JSON.stringify(routerBase)};
  var isInternalDirect = ${JSON.stringify(isInternalUrl(redirectUrl))};
  var redirectUrl = ${JSON.stringify(redirectUrl)};
  var countdown = el 
  ? (isIE ? el.innerText : el.textContent)
  : 0;
  
  function redirect() { 
    if (countdown <= 0) { 
      if (isInternalDirect) {
         redirectUrl = routerBase + redirectUrl.replace(/^\\//, '')
      }
      location.href = redirectUrl

    } else {
      if (isIE) { 
        el.innerText = countdown--; 
      } else { 
        el.textContent = countdown--; 
      } 
    } 
  } 

  redirect();
  setInterval(redirect, 1000); 
  </script>
</body>
</html>`
}
