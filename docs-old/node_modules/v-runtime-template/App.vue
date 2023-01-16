<template>
  <div>
    <v-runtime-template :template="testMixin"/>
    <v-runtime-template :template="testLocal"/>
  </div>
</template>

<script>
import VRuntimeTemplate from "./index.js"; //import raw src
import Test from "./Test.vue"; //import raw src

export default {
  data() {
    return {
      testMixin:
        "<test>Testing Mixins: <ul><li>{{testingProp}}</li><li>{{testingData}}</li><li>{{testingMethod('x')}}</li><li>{{testingComputed}}</li></ul></test>",
      testLocal:
        "<test>Testing Local: <ul><li>{{testingPropLocal}}</li><li>{{testingDataLocal}}</li><li>{{testingMethodLocal('x')}}</li><li>{{testingComputedLocal}}</li></ul></test>",
      testingDataLocal: "localTest: testingData"
    };
  },
  props: {
    testingProp: {
      default: "should not see"
    },
    testingPropLocal: {
      default: "localTest: testingProp"
    }
  },
  computed: {
    testingComputedLocal() {
      return "localTest: testingComputed";
    },
    testingComputed() {
      return "should not see";
    }
  },
  methods: {
    testingMethodLocal() {
      return "localTest: testingMethod";
    },
    testingMethod() {
      return "Should not see";
    }
  },
  components: {
    VRuntimeTemplate,
    // eslint-disable-next-line
    Test
  }
};
</script>

<style>
</style>
