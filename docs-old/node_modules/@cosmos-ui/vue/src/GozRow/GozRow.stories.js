import GozRow from "./GozRow.vue";
import IconIbc from "../Icons/IconIbc"

export default {
  title: "GozRow",
  component: GozRow
};

export const normal = () => ({
  components: {
    GozRow,
    IconIbc
  },
  template: `
    <div style="background: #151831; padding: 1rem 0;">
      <GozRow logo="ibc" url="https://google.com">
        <template v-slot:icon>
          <icon-ibc/>
        </template>
        <template v-slot:h1>
          Demoing the Relayer
        </template>
        <template v-slot:h2>
          cosmos/relayer
        </template>
        <template v-slot:body>
          An example of a server side IBC relayer to be used for Game of Zones and beyond.
        </template>
      </GozRow>
    </div>
  `
});