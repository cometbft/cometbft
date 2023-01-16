<template>
  <div>
    <div class="container">
      <section-input
        :value="query"
        @input="querySet($event)"
        @keypress="inputKeypress"
        @cancel="$emit('cancel', false)"
      />
      <div class="results">
        <section-results-list
          v-if="query && resultsAvailable"
          @activate="$emit('select', $event)"
          :selected="selectedIndex"
          :value="results"
          :base="base"
        />
        <section-shortcuts v-else-if="!query"/>
        <section-results-empty
          v-else-if="query && !resultsAvailable && !searchInFlight"
          @query="querySet($event)"
          :query="query"
        />
      </div>
    </div>
  </div>
</template>

<style scoped>
strong {
  font-weight: 500;
}
.container {
  height: 100vh;
  overflow-y: scroll;
  -webkit-overflow-scrolling: touch;
  flex-direction: column;
  background-color: #f8f9fc;
  font-family: var(--ds-font-family);
}
.results {
  padding-bottom: 3rem;
  display: flex;
  flex-direction: column;
  flex-grow: 1;
}
</style>

<script>
import Fuse from "fuse.js";
import MarkdownIt from "markdown-it";
import hotkeys from "hotkeys-js";
import SectionShortcuts from "./SectionShortcuts.vue";
import SectionResultsEmpty from "./SectionResultsEmpty.vue";
import SectionInput from "./SectionInput.vue";
import SectionResultsList from "./SectionResultsList.vue";
import algoliasearch from "algoliasearch";

const algoliaInit = config => {
  if (config && config.id && config.key && config.index) {
    const algoliaClient = algoliasearch(config.id, config.key);
    const algolia = algoliaClient.initIndex(config.index);
    return algolia;
  } else {
    return false;
  }
};

const fuseInit = site => {
  return new Fuse(
    site.pages
      .map(doc => {
        return {
          key: doc.key,
          title: doc.title,
          headers: doc.headers && doc.headers.map(h => h.title).join(" "),
          description: doc.frontmatter.description,
          path: doc.path
        };
      })
      .filter(doc => {
        return !(
          Object.keys(site.locales || {}).indexOf(doc.path.split("/")[1]) > -1
        );
      }),
    {
      keys: ["title", "headers", "description", "path"],
      shouldSort: true,
      includeScore: true,
      includeMatches: true
    }
  );
};

const fuseFormat = results => {
  return results.map(result => {
    return {
      title: result.item && result.item.title && md(result.item.title),
      desc:
        result.item && result.item.description && md(result.item.description),
      id: result.item && result.item.key
    };
  });
};

const algoliaFormat = results => {
  return results.map(result => {
    const title = Object.values(result.hierarchy)
      .filter(e => e)
      .map(e => e.replace(/^#/, ""))
      .join(" › ");
    return {
      title,
      desc: result.content,
      url: result.url
    };
  });
};

const md = string => {
  const md = new MarkdownIt({ html: true, linkify: true });
  return `<div>${md.renderInline(string)}</div>`;
};

export default {
  props: {
    query: {
      type: String
    },
    site: {
      type: Object
    },
    algoliaConfig: {
      type: Object
    },
    base: {
      default: "/master/"
    }
  },
  components: {
    SectionInput,
    SectionShortcuts,
    SectionResultsEmpty,
    SectionResultsList
  },
  data: function() {
    return {
      results: null,
      fuse: null,
      algolia: null,
      selectedIndex: null,
      searchInFlight: null
    };
  },
  watch: {
    query() {
      this.search(this.query);
    }
  },
  computed: {
    resultsAvailable() {
      return this.results && this.results.length > 0;
    }
  },
  mounted() {
    hotkeys("down", e => {
      this.inputKeypress(e);
      e.preventDefault();
    });
    hotkeys("up", e => {
      this.inputKeypress(e);
      e.preventDefault();
    });
    hotkeys("enter", e => {
      this.inputKeypress(e);
      e.preventDefault();
    });
    this.algolia = algoliaInit(this.algoliaConfig);
    this.fuse = fuseInit(this.site);
    this.search(this.query);
  },
  methods: {
    querySet(string) {
      this.$emit("query", string);
    },
    inputKeypress(e) {
      if (e.key) {
        if (e.key === "ArrowUp") this.selectResult(-1);
        if (e.key === "ArrowDown") this.selectResult(+1);
        if (e.key === "Enter") {
          this.$emit("select", { ...this.results[this.selectedIndex] });
        }
      }
    },
    selectResult(delta) {
      const index = this.selectedIndex,
        indexNew = index + delta,
        isValidIndex = Number.isInteger(index) && index >= 0;
      if (isValidIndex) {
        this.selectedIndex = indexNew >= 0 ? indexNew : 0;
      } else {
        this.selectedIndex = 0;
      }
    },
    async search(query) {
      this.searchInFlight = true;
      if (!query) return;
      if (this.algolia) {
        const params = { hitsPerPage: 100 };
        const results = (await this.algolia.search(query, params)).hits;
        this.results = algoliaFormat(results);
        this.searchInFlight = false;
      } else {
        this.results = fuseFormat(this.fuse.search(query));
      }
    }
  }
};
</script>