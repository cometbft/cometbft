<template>
  <div>
    <div class="row" v-for="item in milestoneList" :key="item.url">
      <goz-row :url="item.url" :progress="item.progress">
        <template v-slot:icon>
          <component :is="`icon-${item.logo}`"/>
        </template>
        <template v-slot:h1>
          {{item.title}}
        </template>
        <template v-slot:h2>
          {{item.repo}}
        </template>
      </goz-row>
    </div>
  </div>
</template>

<style scoped>
.row {
  margin-bottom: 1rem;
}
.row:last-child {
  margin-bottom: initial;
}
</style>

<script>
import GozRow from "../GozRow/GozRow"
import IconIbc from "../Icons/IconIbc"
import IconSdk from "../Icons/IconSdk"
import axios from "axios"

export default {
  components: {
    GozRow,
    IconIbc,
    IconSdk
  },
  data: function() {
    return {
      milestone: {
        logo: "sdk",
        h1: "Cosmos SDK - GoZ Milestone",
        h2: "cosmos/sdk"
      },
      milestoneList: [],
      sources: [
        ["cosmos/cosmos-sdk", 24, "sdk", "Cosmos SDK – GoZ Milestone"],
        ["cosmos/cosmos-sdk", 21, "sdk", "Cosmos SDK – IBC 1.0 Milestone"],
        ["cosmos/relayer", 2, "ibc", "Relayer – GoZ Milestone"],
        ["cosmos/ics", 5, "ibc", "IBC 1.0 – Spec Milestone"],
      ],
    }
  },
  async mounted() {
    this.sources.forEach(async source => {
      const milestone = await this.getMilestone.apply(null, source)
      this.milestoneList.push(milestone)
    })
  },
  methods: {
    async getMilestone(repo, id, logo, defaultTitle, defaultProgress) {
      const url = `https://github.com/${repo}/milestone/${id}`
      try {
        const
          api = `https://api.github.com/repos/${repo}/milestones/${id}`,
          m = (await axios.get(api)).data,
          title = m.title,
          open = parseInt(m.open_issues),
          closed = parseInt(m.closed_issues),
          progress = Math.floor((100 * closed) / (open + closed)).toFixed(0)
        return { title, repo, progress, logo, url }
      } catch {
        return {
          repo,
          logo,
          url,
          title: defaultTitle,
          progress: null,
        }
      }
    }
  }
}
</script>