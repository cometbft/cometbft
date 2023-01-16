<template>
  <div>
    <navbar v-bind="{items}" @selected="selected = $event">
      <template v-slot:dropdown>
        <div v-if="selected === items[0]">Item 1 Content</div>
        <div v-if="selected === items[1]">Item 2 Content</div>
        <div v-if="selected === items[2]">Item 3 Content</div>
      </template>
    </navbar>
  </div>
</template>

<script>
import Navbar from "./Navbar.vue"

export default {
  components: { Navbar },
  data: function() {
    return {
      selected: null,
      items: [
        "Item 1",
        "Item 2",
        "Item 3"
      ]
    }
  }
}
</script>