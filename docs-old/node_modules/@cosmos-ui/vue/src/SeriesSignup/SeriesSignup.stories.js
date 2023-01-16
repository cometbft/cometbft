import SeriesSignup from "./SeriesSignup"
import { text } from '@storybook/addon-knobs';

export default {
  title: "SeriesSignup",
  component: SeriesSignup
}

export const normal = () => ({
  components: { SeriesSignup },
  template: `
    <series-signup>
      <template v-slot:h1>
        Sign up for Code with Us
      </template>
    </series-signup>
  `
})