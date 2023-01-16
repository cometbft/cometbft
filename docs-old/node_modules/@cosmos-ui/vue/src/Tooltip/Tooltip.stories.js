import Tooltip from "./Tooltip.vue";

export default {
  title: "Tooltip",
  component: Tooltip
};

export const normal = () => ({
  components: { Tooltip },
  template: `
    <span>This is an example of <tooltip value="blockchain">blockchain</tooltip> and <tooltip value="consensus-algorithm">consensus algorithm</tooltip></span>
  `
});
