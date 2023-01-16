import GozSection from "./GozSection";
import IconPlanet from "./../Icons/IconPlanet";
import IconWorkout from "./../Icons/IconWorkout";

export default {
  title: "GozSection",
  component: GozSection
};

export const normal = () => ({
  components: { GozSection, IconPlanet },
  template: `
    <div>
      <goz-section>
        <template v-slot:icon>
          <icon-planet/>
        </template>
        <template v-slot:title>
          Coming to a cosmos near you
        </template>
        <template v-slot:subtitle>
          The Cosmos Game of Zones will officially start before summer 2020. A more exact start date will be announced when the IBC demo for GoZ nears readiness. Keep an eye on the milestones on GitHub to track progress:
        </template>
      </goz-section>
    </div>
  `
});

export const notes = () => ({
  components: { GozSection, IconWorkout },
  template: `
    <div>
      <goz-section>
        <template v-slot:icon>
          <icon-calendar/>
        </template>
        <template v-slot:title>
          Preparing for Game of Zones
        </template>
        <template v-slot:notes>
          <ul>
            <li>Expect to register with a Cosmos address and a chain-id for your team.</li>
            <li>You will need to be able to run a blockchain with your team.</li>
            <ul>
              <li>Familiarize yourself with the <a href="https://github.com/cosmos/cosmos-sdk/branches/all?utf8=%E2%9C%93&query=ibc" target="_blank" rel="noreferrer noopener">IBC branches</a> of the Cosmos SDK on GitHub.</li>
              <li class="note"><b>Note:</b> these branches will evolve over time.</li>
            </ul>
            <li>Practice sending IBC transactions between testnets with the Relayer. Follow the <a href="https://cosmos.network" target="_blank" rel="noreferrer noopener">step-by-step instructions</a> to run the end-to-end handshake & token transfer demo.</li>
          </ul>
        </template>
      </goz-section>
    </div>
  `
});