import Stars from "./Stars.vue";
import { withKnobs, text } from '@storybook/addon-knobs';

export default {
  title: "Stars",
  component: Stars,
  decorators: [withKnobs]
};

export const normal = () => ({
  components: { Stars },
  props: {
    vertical: {
      default: text("Vertical", "20")
    },
    horizontal: {
      default: text("Horizontal", ".25")
    },
    flickering: {
      default: text("Flickering", "600")
    }
  },
  template: `
    <div>
      <stars v-bind="{vertical: parseInt(vertical), horizontal: parseFloat(horizontal), flickering: parseInt(flickering)}"/>
    </div>
  `
});