<template lang="pug">
  div
    .container
      .search-box
        .search-box__icon
          icon-search(:stroke="query ? 'var(--color-link)' : '#aaa'" :fill="query ? 'var(--color-link)' : '#aaa'")
        .search-box__input
          input(type="text" autocomplete="off" placeholder="Search" id="search-box-input" ref="search" :value="query" @input="$emit('query', $event.target.value)").search-box__input__input
        .search-box__clear
          icon-circle-cross(v-if="query && query.length > 0" @click.native="$emit('query', '')" @keydown.enter="$emit('query', '')" tabindex="1").search-box__clear__icon
        a.search-box__button(@click="$emit('visible', false)" @keydown.enter="$emit('visible', false)" tabindex="1") Cancel
      .results
        .shortcuts(v-if="!query")
          .shortcuts__h1 Keyboard shortcuts
          .shortcuts__table
            .shortcuts__table__row
              .shortcuts__table__row__keys
                .shortcuts__table__row__keys__item /
              .shortcuts__table__row__desc Open search window
            .shortcuts__table__row
              .shortcuts__table__row__keys
                .shortcuts__table__row__keys__item(style="font-size: .65rem") esc
              .shortcuts__table__row__desc Close search window
            .shortcuts__table__row
              .shortcuts__table__row__keys
                .shortcuts__table__row__keys__item ↵
              .shortcuts__table__row__desc Open highlighted search result
            .shortcuts__table__row
              .shortcuts__table__row__keys
                .shortcuts__table__row__keys__item(style="font-size: .65rem") ▼
                .shortcuts__table__row__keys__item(style="font-size: .65rem") ▲
              .shortcuts__table__row__desc Navigate between search results
        .results__noresults__container(v-if="query && (searchResults && searchResults.length <= 0)")
          .results__noresults
            .results__noresults__icon
              icon-search
            .results__noresults__h1 No results for #[strong “{{query}}”]
            .results__noresults__p
              span Try queries such as #[span.results__noresults__a(@click="query = 'auth'" @keydown.enter="query = 'auth'" tabindex="0") auth], #[span.results__noresults__a(@click="query = 'slashing'" @keydown.enter="query = 'slashing'" tabindex="0") slashing], or #[span.results__noresults__a(@click="query = 'staking'" @keydown.enter="query = 'staking'" tabindex="0") staking].
        div(v-if="query && searchResults && searchResults.length > 0")
          .results__item(@keydown.40="focusNext" @keydown.38="focusPrev" tabindex="0" ref="result" v-for="result in searchResults" v-if="searchResults" @keydown.enter="itemClick(resultLink(result), result.item)" @click="itemClick(resultLink(result), result.item)")
            .results__item__title(v-html="resultTitle(result)")
            .results__item__desc(v-if="resultSynopsis(result)" v-html="resultSynopsis(result)")
            .results__item__h2(v-if="resultHeader(result)") {{resultHeader(result).title}}
</template>

<style lang="stylus" scoped>
.shortcuts
  display flex
  justify-content center
  flex-direction column

  &__h1
    text-align center
    color rgba(22, 25, 49, 0.65)
    text-transform uppercase
    letter-spacing 0.2em
    font-size 0.75rem
    margin-top 4rem
    margin-bottom 2rem

  &__table
    &__row
      display grid
      grid-template-columns 3fr 7fr
      gap 1.5rem
      align-items center
      margin-top 0.5rem
      margin-bottom 0.5rem

      &__keys
        display flex
        justify-content flex-end

        &__item
          color #46509F
          background-color rgba(176, 180, 207, 0.2)
          border 1px solid rgba(176, 180, 207, 0.09)
          border-radius 0.25rem
          font-size 0.8125rem
          width 1.5rem
          height 1.5rem
          display flex
          align-items center
          justify-content center
          margin 2px

      &__desc
        color rgba(22, 25, 49, 0.65)

strong
  font-weight 600

.container
  height 100vh
  overflow-y scroll
  -webkit-overflow-scrolling touch
  // display flex
  flex-direction column
  background-color #F8F9FC

.search-box
  width 100%
  display grid
  grid-auto-flow column
  align-items center
  box-shadow inset 0 -1px 0 0 rgba(176, 180, 207, 0.2)
  padding-left 1.5rem
  padding-right 1.5rem
  grid-template-columns 2rem 1fr 1.25rem
  gap 0.5rem

  &__icon
    margin-left 0.5rem

  &__input
    &__input
      border none
      background none
      outline none
      font-size 1.25rem
      width 100%
      padding 1.5rem 0.5rem

      &::-webkit-input-placeholder
        color rgba(0, 0, 0, 0.46)

      &:hover::-webkit-input-placeholder
        color rgba(0, 0, 0, 0.67)

      &:focus::-webkit-input-placeholder
        color rgba(0, 0, 0, 0.46)

  &__clear
    &__icon
      cursor pointer
      fill rgba(0, 0, 0, 0.15)
      margin-top 0.25rem

      &:hover, &:focus
        fill rgba(0, 0, 0, 0.25)
        outline none

      &:active
        opacity 0.7

  &__button
    text-transform uppercase
    color var(--color-link)
    font-weight 600
    cursor pointer
    height 100%
    display flex
    padding-left 0.5rem
    padding-right 0.5rem
    align-items center
    outline 0
    transition opacity .2s ease-out

    &:focus
      background-color rgba(176, 180, 207, 0.087)

    &:active
      opacity 0.7
      transition-duration 0s

.results
  padding-bottom 3rem
  display flex
  flex-direction column
  flex-grow 1

  &__noresults__container
    height 100%
    display flex
    flex-direction column
    align-items center
    justify-content center
    flex-grow 1
    margin-top 4rem

  &__noresults
    display flex
    flex-direction column
    align-items center
    justify-content center

    &__icon
      max-width 80px
      margin-bottom 2rem
      fill #ccc

    &__h1
      color rgba(22, 25, 49, 0.65)
      font-size 1.5rem
      margin-bottom 1rem

    &__p
      color rgba(22, 25, 49, 0.65)

    &__a
      cursor pointer
      color var(--color-link)

  &__item
    padding 0.75rem 2rem
    cursor pointer

    &:hover,
    &:focus
      outline none
      background-color rgba(176, 180, 207, 0.087)

    &__title
      color var(--color-link)
      line-height 1.5rem

    &__h2
      margin-top 0.25rem
      margin-bottom 0.25rem
      font-weight 600
      font-size 0.875rem
      line-height 1.25rem

      &__item
        display inline-block

        &:after
          content '>'
          margin-left 0.25rem
          margin-right 0.25rem

        &:last-child
          &:after
            content ''

    &__desc
      opacity 0.5
      white-space nowrap
      overflow hidden
      position relative
      font-size 0.875rem
      line-height 1.25rem

      &:after
        content ''
        background linear-gradient(to right, transparent 0%, rgba(248, 249, 252, 1))
        height 1em
        width 2.5em
        padding-bottom 0.25rem
        text-align right
        position absolute
        top 0
        right 0

@media screen and (max-width 768px)
  .search-box
    padding-left 1rem
    padding-right 1rem

  .shortcuts
    display none
</style>

<script>
import { find, last, debounce } from "lodash";
//- import Fuse from "fuse.js";

export default {
  props: ["visible", "query"],
  data: function() {
    return {
      searchResults: null,
      //- searchQuery: null,
      //- fuse: null,
    };
  },
  watch: {
    query: function(e) {
      return this.debouncedSearch();
    },
    visible(becomesVisible) {
      const search = this.$refs.search;
      if (becomesVisible && search) {
        search.select();
      }
    },
  },
  computed: {
    debouncedSearch() {
      return debounce(this.search, 300);
    },
  },
  mounted() {
    this.$refs.search.addEventListener("keydown", (e) => {
      if (e.keyCode == 27) {
        this.$emit("visible", false);
        return;
      }
      if (e.keyCode == 40) {
        this.$refs.result[0].focus();
        e.preventDefault();
        return;
      }
    });
    //- const fuseIndex = this.$site.pages
    //-   .map((doc) => {
    //-     return {
    //-       key: doc.key,
    //-       title: doc.title,
    //-       headers: doc.headers && doc.headers.map((h) => h.title).join(" "),
    //-       // description: doc.frontmatter && doc.frontmatter.description,
    //-       path: doc.path,
    //-     };
    //-   })
    //-   .filter((doc) => {
    //-     return !(
    //-       Object.keys(this.$site.locales || {}).indexOf(
    //-         doc.path.split("/")[1]
    //-       ) > -1
    //-     );
    //-   });
    //- const fuseOptions = {
    //-   keys: ["title", "headers", "description", "path"],
    //-   shouldSort: true,
    //-   includeScore: true,
    //-   includeMatches: true,
    //-   threshold: 1,
    //- };
    //- this.fuse = new Fuse(fuseIndex, fuseOptions);
    if (this.$refs.search) this.$refs.search.focus();
    //- this.search();
  },
  methods: {
    resultTitle(result) {
      const path = this.itemPath(result.item)
        ? this.itemPath(result.item) + " /"
        : "";
      return this.md(`${path} ${result.item.title}`);
    },
    resultSynopsis(result) {
      if (!result.item.frontmatter.description) return false;
      return this.md(
        result.item.frontmatter.description
          .split("")
          .slice(0, 75)
          .join("") + "..."
      );
    },
    resultLink(result) {
      const header = this.resultHeader(result);
      return result.item.path + (header ? `#${header.slug}` : "");
    },
    resultHeader(result) {
      if (!result.item.headers) return false;
      const headers = result.item.headers.filter((h) =>
        h.title.match(new RegExp(this.query, "gi"))
      );
      if (headers && headers.length) return headers[0];
    },
    //- search(e) {
    //-   if (!this.query) return;
    //-   const fuse = this.fuse.search(this.query).map((result) => {
    //-     return {
    //-       ...result,
    //-       item: find(this.$site.pages, { key: result.item.key }),
    //-     };
    //-   });
    //-   this.searchResults = fuse;
    //- },
    itemByKey(key) {
      return find(this.$site.pages, { key });
    },
    itemSynopsis(item) {
      return (
        this.itemByKey(item.ref) &&
        this.itemByKey(item.ref).frontmatter &&
        this.itemByKey(item.ref).frontmatter.description &&
        this.md(this.itemByKey(item.ref).frontmatter.description)
      );
    },
    itemClick(url, item) {
      this.$emit("visible", false);
      if (item.path != this.$page.path) {
        this.$router.push(url);
      }
    },
    itemPath(sourceItem) {
      let path = sourceItem.path
        .split("/")
        .filter((item) => item !== "")
        .map((currentValue, index, array) => {
          let path = array.slice(0, index + 1).join("/");
          return "/" + path;
        })
        .map((item) => {
          return /\.html$/.test(item) ? item : `${item}/`;
        });
      path = path.map((item) => {
        const found = find(this.$site.pages, (page) => {
          return page.regularPath === item;
        });
        const noIndex = {
          title: last(item.split("/").filter((e) => e !== "")),
          path: "",
        };
        return found ? found : noIndex;
      });
      return path
        .map((p) => p.title)
        .slice(0, -1)
        .pop();
    },
    focusNext(e) {
      const next = e.target.nextSibling;
      if (next && next.focus) next.focus();
      e.preventDefault();
    },
    focusPrev(e) {
      const prev = e.target.previousSibling;
      if (prev && prev.focus) prev.focus();
      e.preventDefault();
    },
    // resultHeader(result) {
    //   if (!result.headers) return;
    //   return result.headers
    //     .map(h => {
    //       if (h.title.match(new RegExp(this.searchQuery, "gi"))) {
    //         return h.title;
    //       }
    //     })
    //     .filter(e => e);
    // }
  },
};
</script>
