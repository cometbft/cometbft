<template>
  <div>
    <button @click="visible = !visible">Open sidebar</button>
    <modal-sheet v-bind="{...modalProps}" :visible="visible" @visible="visible = $event">
      <section-search v-bind="{algoliaConfig, query, site}" @cancel="visible = false" @query="query = $event"/>
    </modal-sheet>
  </div>
</template>

<script>
import Modal from "../Modal/Modal.vue"
import SectionSearch from "./SectionSearch.vue"
import { default as site } from "./site.js"

export default {
  props: {
    algoliaConfig: {
      type: Object
    }
  },
  components: {
    SectionSearch,
    "modal-sheet": Modal
  },
  data: function() {
    return {
      visible: null,
      query: null,
      modalProps: {
        backgroundColor: 'transparent',
        boxShadow: '0 0 50px rgba(0,0,0,.1)',
        side: "right",
        maxWidth: "400px",
        width: "100%"
      },
      site
    }
  },
}
</script>