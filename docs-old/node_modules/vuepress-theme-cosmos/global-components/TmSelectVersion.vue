<template lang="pug">
  div
    .container
      span.sr-only Docs Version Switcher
      .select(v-if="versions")
        select(@input="versionChange($event.target.value)" v-model="selectedItem")
          option(value="" disabled) Version
          option(v-for="item in versions" :value="item.key") {{ item.label }}
</template>

<script>
export default {
  data() {
    return {
      selectedItem: "",
    }
  },
  mounted() {
    const emptyVal = ""
    // #91: remove trailing slash to fix cloudfront, netlify, gh-pages's deployed URLs
    const pathName = window.location.pathname.replace("/", "").replace(/\/$/, "")
    const pathRe = /[a-zA-Z]{1}\d+(\.\d+)+/g
    const versionPathName = window.location.href.match(pathRe)

    // match values in select option
    // input: https://docs.cosmos.network/master
    // output: null
    if (this.versionValue === this.versionValue) {
      // check if input has a pathname
      // extract version number from string
      // point to the selectedItem
      // input: https://docs.cosmos.network/v0.39/intro/overview.html
      // output: [ 'v0.39' ]
      if (versionPathName !== null) {
        this.selectedItem = versionPathName[0]
      } else {
        this.selectedItem = pathName
      }
    }
    // input: http://docs.cosmos.network
    // expected: 'Version'
    else {
      this.selectedItem = emptyVal
    }
  },
  computed: {
    versions() {
      return this.$themeConfig.versions;
    },
    versionValue() {
      for (var key in this.versions) {
        var value = this.versions[key];

        return value.key
      }
    }
  },
  methods: {
    versionChange(version) {
      // vue router won't work because of the generated path prefix in makefile
      // this.$router.push({ path: `/${version}` }, () => {})
      // to fix urls with path prefixes: https://docs.staging-cosmos.network/master/master
      // window.open(`${window.location.origin}/${version}`)
      window.location.href = `${window.location.origin}/${version}`
    }
  }
};
</script>

<style lang="stylus" scoped>
// Accessible/SEO friendly CSS hiding
.sr-only
  position absolute
  height 1px
  width 1px
  overflow hidden
  clip rect(1px, 1px, 1px, 1px)

select
  border none
  letter-spacing 0.03em
  font-weight 600
  font-size 0.875rem
  line-height 1.25rem
  color rgba(0, 0, 0, 0.667)
  padding 0.5rem 0
  max-width 100%
  box-sizing border-box
  appearance none
  background none
  background-image url("data:image/svg+xml,%3Csvg width='16' height='16' viewBox='0 0 16 16' fill='none' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='M2.5 5L8 10.5L13.5 5' stroke='black' stroke-opacity='0.667' stroke-width='2' stroke-linecap='round'/%3E%3C/svg%3E%0A")
  background-repeat no-repeat, repeat
  background-position right .7em top 50%, 0 0
  background-size .75em auto, 100%
  padding-right 1.75rem

  &:focus
    outline none

  &:hover
    cursor pointer

.select
  background-color transparent
  width fit-content
</style>
