import { withKnobs, text, select, boolean } from '@storybook/addon-knobs';
import ShowcaseButton from "./ShowcaseButton.vue"

export default {
  title: "Button",
  decorators: [withKnobs]
};

export const showcase = () => ({
  components: { ShowcaseButton },
  template: `
    <div>
      <showcase-button/>
    </div>
  `
});
