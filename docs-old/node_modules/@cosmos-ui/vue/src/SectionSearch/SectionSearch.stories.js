import SectionSearch from "./SectionSearch.vue";
import SectionResultsEmpty from "./SectionResultsEmpty.vue";
import SectionShortcuts from "./SectionShortcuts.vue";
import SectionInput from "./SectionInput.vue";
import ModalSearchShowcase from "./ModalSearchShowcase.vue";
import { default as site } from "./site.js";

export default {
  title: "SectionSearch",
  component: SectionSearch,
};

const algoliaConfig = {
  id: "BH4D9OD16A",
  key: "ac317234e6a42074175369b2f42e9754",
  index: "cosmos-sdk",
};

export const normal = () => ({
  components: {
    SectionSearch,
    SectionResultsEmpty,
    SectionShortcuts,
    SectionInput,
  },
  data: function() {
    return {
      site,
      query: null,
      algoliaConfig,
    };
  },
  methods: {
    log: (e) => console.log(e),
  },
  template: `
    <div style="width: 100%; max-width: 600px">
      <section-search v-bind="{algoliaConfig, query, site}" @select="log($event)" @cancel="log('cancel')" @query="query = $event"/>
      <p>Input component:</p>
      <section-input :value="query" @input="query = $event" style="background: #f8f9fc"/>
      <p>Shortcuts section:</p>
      <section-shortcuts style="background: #f8f9fc"/>
      <p>No results section:</p>
      <section-results-empty :query="query" @query="query = $event" style="background: #f8f9fc"/>
    </div>
  `,
});

export const modalSearch = () => ({
  components: {
    ModalSearchShowcase,
  },
  data: function() {
    return {
      algoliaConfig,
    };
  },
  template: `
    <div>
      <modal-search-showcase v-bind="{algoliaConfig}"/>
    </div>
  `,
});
