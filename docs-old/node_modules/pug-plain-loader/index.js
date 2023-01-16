const pug = require('pug')
const loaderUtils = require('loader-utils')

module.exports = function (source) {
  const options = Object.assign({
    filename: this.resourcePath,
    doctype: 'html',
    compileDebug: this.debug || false
  }, loaderUtils.getOptions(this))

  const template = pug.compile(source, options)
  template.dependencies.forEach(this.addDependency)
  return template(options.data || {})
}
