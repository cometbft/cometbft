import GozMilestones from "./GozMilestones.vue";

export default {
  title: "GozMilestones",
  component: GozMilestones
};

export const normal = () => ({
  components: { GozMilestones },
  template: `
    <div style="background: #151831; padding: 1rem 0">
      <GozMilestones/>
    </div>
  `
});