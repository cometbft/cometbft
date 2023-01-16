import GozCard from "./GozCard"
import { text } from '@storybook/addon-knobs';

export default {
  title: "GozCard",
  component: GozCard
}

export const normal = () => ({
  components: { GozCard },
  props: {
    imgSrc: {
      default: text("Image URL", "/goz.jpg")
    },
  },
  template: `
    <goz-card v-bind="{ imgSrc }" />
  `
})