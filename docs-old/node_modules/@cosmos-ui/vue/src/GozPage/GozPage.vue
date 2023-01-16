<template>
  <div>
    <div class="container">
      <goz-hero />
      <!-- <goz-section>
        <template v-slot:subtitle>
          <p>In preparation for the launch of the forthcoming <a href="https://cosmos.network/ibc">Inter-Blockchain Communication (IBC)</a> module, we are announcing Cosmos's second ever adversarial testnet challenge: Game of Zones.</p>
          <p>This competition will create an invaluable educational opportunity and environment for the Cosmos community to learn the nuances of the IBC protocol, including exception handling and potential failure modes.</p>
          <p>More information on this competition will be coming soon, so stay tuned for updates.</p>
        </template>
      </goz-section> -->
      <goz-section>
        <template v-slot:icon>
          <icon-planet/>
        </template>
        <template v-slot:title>
          Coming to a cosmos near you
        </template>
        <!-- <template v-slot:subtitle>
          <p>The Cosmos Game of Zones will officially start before summer 2020. A more exact start date will be announced when the IBC demo for GoZ nears readiness.</p>
          <p>Keep an eye on the milestones on GitHub to track progress:</p>
        </template> -->
        <template v-slot:subtitle>
          <p>In preparation for the launch of the forthcoming <a href="https://cosmos.network/ibc">Inter-Blockchain Communication (IBC)</a> module, we are announcing Cosmos's second ever adversarial testnet challenge: Game of Zones.</p>
          <p>This competition will create an invaluable educational opportunity and environment for the Cosmos community to learn the nuances of the IBC protocol, including exception handling and potential failure modes.</p>
          <p>To get ahead of the game, you can <a href="https://forms.gle/umjGa9G3Q6hcq2iF7" target="_blank" rel="noreferrer noopener">register</a> and start preparing your team for the first challenge.</p>
          <p>More information on this competition will be coming soon, so stay tuned for updates.</p>
        </template>
      </goz-section>
      <!-- <goz-milestones class="card__narrow"/> -->
      <goz-newsletter-signup class="card" v-bind="{ imgSrc }" />
      <goz-section>
        <template v-slot:icon>
          <icon-calendar />
        </template>
        <template v-slot:title>
          Schedule
        </template>
      </goz-section>
      <div style="max-width: 44rem; margin: 0 auto;">
        <goz-schedule :value="value"/>
      </div>
      <goz-section>
        <template v-slot:icon>
          <icon-cup/>
        </template>
        <template v-slot:title>
          One competition, many winners
        </template>
        <template v-slot:subtitle>
          <p>The Interchain Foundation is providing <strong>100,000 ATOM</strong> for the prize pool. Unlike last year's <i>Game of Stakes</i>, there won't be just one winner, and there won't be just one prize. There will be multiple prizes attached to multiple challenges.</p>
        </template>
      </goz-section>
      <goz-section>
        <template v-slot:icon>
          <icon-calendar/>
        </template>
        <template v-slot:title>
          Preparing for Game of Zones
        </template>
        <template v-slot:notes>
          <ul>
            <li><a href="https://forms.gle/umjGa9G3Q6hcq2iF7" target="_blank" rel="noreferrer noopener">Register</a> with a Cosmos address and a chain-id for your team.</li>
            <li>You will need to be able to run a blockchain with your team.</li>
            <ul>
              <li>Familiarize yourself with the <a href="https://github.com/cosmos/cosmos-sdk/branches/all?utf8=%E2%9C%93&query=ibc" target="_blank" rel="noreferrer noopener">IBC branches</a> of the Cosmos SDK on GitHub.</li>
              <li class="note"><b>Note:</b> these branches will evolve over time.</li>
            </ul>
            <li>Practice sending IBC transactions between testnets with the Relayer. Follow the <a href="https://github.com/iqlusioninc/relayer#demoing-the-relayer" target="_blank" rel="noreferrer noopener">step-by-step instructions</a> to run the end-to-end handshake & token transfer demo.</li>
          </ul>
        </template>
      </goz-section>
      <goz-row class="card__narrow" url="https://github.com/iqlusioninc/relayer">
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
      </goz-row>
      <goz-card class="card" v-bind="{ imgSrc }" />
    </div>
  </div>
</template>

<style scoped>
.container {
  background: var(--grey-14);
}
.card {
  padding-top: 4rem;
  padding-bottom: 7rem;
  padding-left: 1rem;
  padding-right: 1rem;
  max-width: 74rem;
  margin-left: auto;
  margin-right: auto;
}
.card__narrow {
  max-width: 49rem;
  margin-left: auto;
  margin-right: auto;
}
@media screen and (max-width: 500px) {
  .card {
    padding-bottom: 2rem;
  }
}
</style>

<script>
import GozHero from "../Hero/GozHero"
import GozCard from "../GozCard/GozCard"
// import GozMilestones from "../GozMilestones/GozMilestones"
import SeriesSignup from "../SeriesSignup/SeriesSignup"
import GozSchedule from "../GozSchedule/GozSchedule"
import GozSection from "../GozSection/GozSection"
import GozRow from "../GozRow/GozRow"
import IconPlanet from "../Icons/IconPlanet"
import IconCalendar from "../Icons/IconCalendar"
import IconWorkout from "../Icons/IconWorkout"
import IconCup from "../Icons/IconCup"
import IconIbc from "../Icons/IconIbc"

export default {
  components: {
    GozHero,
    GozCard,
    // GozMilestones,
    SeriesSignup,
    GozSchedule,
    GozSection,
    GozRow,
    IconPlanet,
    IconCalendar,
    IconWorkout,
    IconCup,
    IconIbc
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
  props: {
    imgSrc: {
      type: String,
      default: "/goz.jpg"
    }
  }
}
</script>