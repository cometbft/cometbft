import Footer from "./Footer.vue";
import { default as data } from "./data"

export default {
  title: "Footer",
  component: Footer
};

export const normal = () => ({
  components: { Footer },
  data: function () {
    return {
      data
    };
  },
  template: `
    <div>
      <Footer v-bind="data" style="padding: 2rem"/>
    </div>
  `
});