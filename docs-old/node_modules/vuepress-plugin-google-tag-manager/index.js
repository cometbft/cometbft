const {path} = require('@vuepress/shared-utils');

module.exports = (options = {}, context) => ({
    define() {
        const {siteConfig = {}} = context;
        const gtm = options.gtm || siteConfig.gtm;
        const GTM_ID = gtm || false;
        return {GTM_ID}
    },


    enhanceAppFiles: [
        path.resolve(__dirname, 'inject.js')
    ],
});