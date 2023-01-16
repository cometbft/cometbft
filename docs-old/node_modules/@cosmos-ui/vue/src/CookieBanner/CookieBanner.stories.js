import CookieBanner from "./CookieBanner.vue";
import { remove } from 'tiny-cookie';

export default {
  title: "CookieBanner",
  component: CookieBanner
};

export const normal = () => ({
  components: { CookieBanner },
  data: function () {
    return {
      banner: true
    };
  },
  template: `
    <div>
      <button @click="reset">Reset the banner</button>
      <cookie-banner />
    </div>
  `,
  methods: {
    reset() {
      remove('cookie-consent-accepted')
      location.reload()
    }
  }
});
