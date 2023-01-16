// vue.config.js

module.exports = {
  runtimeCompiler: true,
  configureWebpack: {
    resolve: {
      alias: {
        vue$: "vue/dist/vue.common"
      }
    }
  }
};