import GozSchedule from "./GozSchedule.vue";

export default {
  title: "GozSchedule",
  component: GozSchedule
};

export const normal = () => ({
  components: {
    GozSchedule
  },
  data: function () {
    return {
      value: [
        {
          date: "Fri, April 24",
          body: "Week 1 Challenge Announcement"
        },
        {
          date: "Sat, April 25",
          body: "[Registration](https://forms.gle/umjGa9G3Q6hcq2iF7) for Game of Zones closes at 11:59PM (PST)"
        },
        {
          date: "Fri, May 1",
          body: `
**Game of Zones launches**
* Official GoZ Opening Ceremonies Live Stream
* Week 2 Challenge Announcement
          `
        },
        {
          date: "Fri, May 8",
          body: "Week 3 Challenge Announcement"
        },
        {
          date: "Fri, May 22",
          body: "**Game of Zones closes**"
        },
        {
          date: "Thu, May 28",
          body: "Official GoZ Closing Ceremonies Live Stream"
        },
      ]
    }
  },
  template: `
    <div style="background: #151831; padding: 1rem 0">
      <div style="max-width: 44rem; margin: 0 auto;">
        <goz-schedule :value="value"/>
      </div>
    </div>
  `
});