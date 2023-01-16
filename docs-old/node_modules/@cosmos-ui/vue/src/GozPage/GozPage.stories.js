import GozPage from "./GozPage"

export default {
  title: "GozPage",
  component: GozPage
}

export const normal = () => ({
  components: { GozPage },
  template: `
    <goz-page/>
  `
})