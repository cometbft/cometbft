<template>
  <div>
    <div class="container">
      <div v-for="(item, i) in value" :key="i" class="row">
        <div class="row__aside">
          {{item.date}}
        </div>
        <div class="row__body" v-html="md(item.body)"></div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.container {
  color: white;
  font-family: var(--ds-font-family, inherit);
  margin-left: 2rem;
  margin-right: 2rem;
}
.row {
  display: grid;
  grid-template-columns: 25% 1fr;
  gap: 2rem;
}
.row__aside {
  text-align: right;
  font-size: .75rem;
  text-transform: uppercase;
  letter-spacing: 0.2em;
  color: rgba(255, 255, 255, 0.51);
  padding-top: 1.5rem;
  padding-bottom: 1.5rem;
  margin-top: .3rem;
}
.row__body {
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  padding-top: 1.5rem;
  padding-bottom: 1.5rem;
  line-height: 1.5;
}
::v-deep{
  font-size: 1rem;
}
::v-deep p {
  margin: 0;
}
::v-deep a {
  color: #66A1FF;
  text-decoration: none;
}
::v-deep ul {
  padding-left: 2em;
}
::v-deep li {
  margin-top: .75rem;
  margin-bottom: .75rem;
  list-style-type: disc;
}
@media screen and (max-width: 600px) {
  .row {
    display: block;
  }
  .row__aside {
    text-align: initial;
    padding-bottom: .5rem;
    font-size: .75rem;
  }
  .row__body {
    padding-top: 0;
  }
  ::v-deep {
    font-size: .875rem;
  }
}
</style>

<script>
import MarkdownIt from "markdown-it"

export default {
  props: {
    value: {
      type: Array
    }
  },
  methods: {
    md(string) {
      const md = new MarkdownIt()
      return md.render(string)
    }
  }
}
</script>