import Hero from "./Hero.vue";
import GozHero from "./GozHero.vue";
import data from "./data"
import { text } from '@storybook/addon-knobs';

export default {
  title: "Hero",
  component: Hero
};

export const normal = () => ({
  components: { Hero },
  data: function () {
    return {
      data: data.hero
    };
  },
  template: `
    <div>
      <div>
        <Hero :value="data" :alt="true"/>
      </div>
    </div>
  `
});

export const wide = () => ({
  components: { Hero },
  data: function () {
    return {
      data: {
        ...data.hero,
        h2: "Zoom out to see the difference",
        background: ""
      }
    };
  },
  template: `
    <div style="background-image: url('https://images.unsplash.com/photo-1531707566548-6577aab321d7?ixlib=rb-1.2.1&ixid=eyJhcHBfaWQiOjEyMDd9&auto=format&fit=crop&w=2850&q=100'); background-size: cover">
      <div style="max-width: 1440px; margin: 0 auto;">
        <Hero :value="data" :alt="true" style="padding: 2rem"/>
      </div>
    </div>
  `
});

export const Goz = () => ({
  components: { GozHero },
  props: {
    imgSrc: {
      default: text("Image URL", "/goz.jpg")
    },
    logoSrc: {
      default: text("Logo URL", "/goz-logo.svg")
    }
  },
  template: `
    <div class="wrapper" style="background-color: var(--grey-14);">
      <goz-hero v-bind="{ imgSrc, logoSrc }" />
    </div>
  `
});