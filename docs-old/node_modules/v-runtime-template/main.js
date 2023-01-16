import Vue from "vue";
import App from "./App.vue";

import testPlugin from "./test.plugin.js"; //testing mixins

Vue.use(testPlugin);

new Vue({
    render: h => h(App)
}).$mount("#app");