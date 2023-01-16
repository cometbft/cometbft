import Cards from "./Cards.vue";
import data from "./data"

export default {
  title: "Cards",
  component: Cards
};

export const normal = () => ({
  components: { Cards },
  data: function () {
    return {
      cards: data.cards
    }
  },
  template: `
    <div>
      <Cards :value="cards"/>
    </div>
  `
});

export const horizontal = () => ({
  components: { Cards },
  data: function () {
    return {
      data: data.cards
    }
  },
  template: `
    <div>
      <Cards :value="data" format="horizontal"/>
    </div>
  `
});
