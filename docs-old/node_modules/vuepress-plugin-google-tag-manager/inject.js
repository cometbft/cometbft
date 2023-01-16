import GtmPlugin from './GtmPlugin'

export default ({router, Vue}) => {

    if (process.env.NODE_ENV === 'production' && GTM_ID && typeof window !== 'undefined') {
        (function (w, d, s, l, i) {
            w[l] = w[l] || [];
            w[l].push({
                'gtm.start':
                    new Date().getTime(), event: 'gtm.js'
            });
            var f = d.getElementsByTagName(s)[0],
                j = d.createElement(s), dl = l != 'dataLayer' ? '&l=' + l : '';
            j.async = true;
            j.src =
                'https://www.googletagmanager.com/gtm.js?id=' + i + dl;
            f.parentNode.insertBefore(j, f);
        })(window, document, 'script', 'dataLayer', GTM_ID);

        Vue.prototype.$gtm = Vue.gtm = new GtmPlugin();

        router.afterEach(function (to) {
            Vue.prototype.$gtm.trackView(to.name, to.fullPath);
        })
    }

}
