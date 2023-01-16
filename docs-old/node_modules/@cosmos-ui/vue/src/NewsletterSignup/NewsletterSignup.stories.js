import CosmosNewsletterSignup from "./CosmosNewsletterSignup.vue";
import { withKnobs, text } from "@storybook/addon-knobs";

export default {
  title: "NewsletterSignup",
  component: CosmosNewsletterSignup,
  decorators: [withKnobs],
};

export const cosmos = () => ({
  props: {
    fullscreen: {
      default: text("fullscreen", "100vh"),
    },
  },
  data() {
    return {
      value: {
        h1: "Sign up for Cosmos updates",
        h2:
          "Get the latest from the Cosmos ecosystem and engineering updates, straight to your inbox.",
        topics: [
          {
            h1: "Tools & technology",
            h2:
              "Engineering and development updates on Cosmos SDK, Tendermint, IBC and more.",
            requestURL: "https://app.mailerlite.com/webforms/submit/o0t6d7",
            callback: "jQuery18307296239382192573_1594158619276",
            _: "1594158625563",
            groups: "103455779",
            svg: "/icon-window-code.svg",
          },
          {
            h1: "Ecosystem & community",
            h2:
              "General news and updates from the Cosmos ecosystem and community.",
            requestURL: "https://app.mailerlite.com/webforms/submit/o0t6d7",
            callback: "jQuery18307296239382192573_1594158619276",
            _: "1594158625563",
            groups: "103455777",
            svg: "/icon-network.svg",
          },
        ],
      },
    };
  },
  components: {
    CosmosNewsletterSignup,
  },
  template: `
    <div>
      <cosmos-newsletter-signup v-bind="{...value, fullscreen}"/>
    </div>
  `,
});

export const ibc = () => ({
  components: {
    CosmosNewsletterSignup,
  },
  data() {
    return {
      value: {
        h1: "Sign up for IBC updates",
        h2:
          "Get engineering, development and ecosystem updates on IBC (Inter-Blockchain Communciation protocol) - straight to your inbox.",
        requestURL: "https://app.mailerlite.com/webforms/submit/y2i9q3",
        callback: "jQuery183003200065485233239_1594158714190",
        _: "1594158730789",
        svg: "/icon-ibc.svg",
        background:
          'url("/stars.svg") repeat, linear-gradient(137.58deg, #161931 9.49%, #2D1731 91.06%)',
      },
    };
  },
  template: `
    <div>
      <cosmos-newsletter-signup v-bind="{...value}"/>
    </div>
  `,
});

export const tools = () => ({
  components: {
    CosmosNewsletterSignup,
  },
  data() {
    return {
      value: {
        h1: "Get Cosmos tools & technology updates",
        h2:
          "Get engineering and development updates on Cosmos SDK, Tendermint Core, IBC and more - straight to your inbox.",
        requestURL: "https://app.mailerlite.com/webforms/submit/u3s3t9",
        callback: "jQuery1830889145586852685_1594158789750",
        _: "1594158795317",
        svg: "/icon-terminal-window.svg",
      },
    };
  },
  template: `
    <div>
      <cosmos-newsletter-signup v-bind="{...value}"/>
    </div>
  `,
});
