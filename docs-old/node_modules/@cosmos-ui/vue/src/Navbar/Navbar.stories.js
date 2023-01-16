import Navbar from "./Navbar.vue";
import ShowcaseNavbar from "./ShowcaseNavbar.vue"
import { withKnobs, text, select, boolean } from '@storybook/addon-knobs';

export default {
  title: "Navbar",
  component: Navbar,
  decorators: [withKnobs]
};

export const normal = () => ({
  components: { Navbar },
  template: `
    <div>
      <navbar/>
    </div>
  `
});

export const showcase = () => ({
  components: { ShowcaseNavbar },
  template: `
    <div>
      <showcase-navbar/>
    </div>
  `
});