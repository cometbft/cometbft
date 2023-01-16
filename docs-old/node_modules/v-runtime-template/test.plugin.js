//https://vuejs.org/v2/guide/plugins.html
//https://dev.to/nkoik/writing-a-very-simple-plugin-in-vuejs---example-8g8
// This exports the plugin object.

import Test from "./Test.vue"

export default {
  // The install method will be called with the Vue constructor as
  // the first argument, along with possible options
  install(Vue) {
    Vue.mixin({
      components:{Test},
      props: {
        testingProp: {
          default: "mixinTest: testingProp"
        }
      },
      data() {
        return {
          testingData: "mixinTest: testingData"
        };
      },
      computed: {
        testingComputed() {
          return "mixinTest: testingComputed";
        }
      },
      methods: {
        testingMethod() {
          return "mixinTest: testingMethod";
        }
      }
    }); //end mixin

    Vue.prototype.$testProto = function (str) {
      return "mixinTest: testingProto=" + str;
    }; //end $testProto

  }
};