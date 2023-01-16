<template>
  <div>
    <transition name="overlay" appear>
      <div class="overlay"
           ref="overlay"
           :style="{'background-color': backgroundColor || 'rgba(0, 0, 0, 0.35)'}"
           v-if="visible && visibleLocal"
           @click="close"
           @touchstart="touchstart"
           @touchmove="touchmove"
           @touchend="touchend">
      </div>
    </transition>
    <transition name="sidebar" @after-leave="$emit('visible', false)" appear>
      <div class="sidebar"
           ref="sidebar"
           v-if="visible && visibleLocal"
           :style="style"
           @touchstart="touchstart"
           @touchmove="touchmove"
           @touchend="touchend">
        <slot/>
      </div>
    </transition>
  </div>
</template>

<style scoped>
.overlay {
  position: fixed;
  top: 0;
  left: 0;
  height: 100vh;
  width: 100vw;
}
.sidebar {
  position: fixed;
  top: 0;
  height: 100vh;
  background: white;
  overflow-y: scroll;
  -webkit-overflow-scrolling: touch;
  transform: translateX(var(--translate-x-component-internal));
}
.overlay-enter-active {
  transition: all .25s ease-in;
}
.overlay-enter {
  opacity: 0;
}
.overlay-enter-to {
  opacity: 1;
}
.overlay-leave-active {
  transition: all .25s;
}
.overlay-leave {
  opacity: 1;
}
.overlay-leave-to {
  opacity: 0;
}
.sidebar-enter-active {
  transition: all .25s;
}
.sidebar-enter {
  transform: translateX(var(--sidebar-transform-component-internal));
}
.sidebar-enter-to {
  transform: translateX(0);
}
.sidebar-leave-active {
  transition: all .25s;
}
.sidebar-leave {
  transform: translateX(0);
}
.sidebar-leave-to {
  transform: translateX(var(--sidebar-transform-component-internal));
}
</style>

<script>
export default {
  props: [
    "visible",
    "width",
    "max-width",
    "side",
    "background-color",
    "box-shadow"
  ],
  data: function() {
    return {
      visibleLocal: true,
      touchStartX: null,
      touchMoveX: 0,
      touchEndX: null
    };
  },
  watch: {
    visible(newValue, oldValue) {
      if (newValue) {
        const body = document.querySelector("body").style;
        const html = document.querySelector("html").style;
        const iOS =
          !!navigator.platform && /iPad|iPhone|iPod/.test(navigator.platform);
        const sidebar = this.$refs.sidebar;
        body.height = "100%"
        body.overflow="hidden"
        html.height = "100%"
        html.overflow="hidden"
        body.overflowY = "hidden";
        // body.overflowX = "hidden";
        if (sidebar) {
          sidebar.addEventListener("transitionend", () => {
            sidebar.style.transition = "";
          });
        }
        this.touchMoveX = null;
        this.touchStartX = null;
        this.visibleLocal = true;
      } else {
        document.querySelector("body").style.overflowY = "";
        document.querySelector("body").style.position = "";
      }
    }
  },
  computed: {
    style() {
      return {
        "box-shadow": this.boxShadow || "none",
        left: this.side === "right" ? "initial" : "0",
        right: this.side === "right" ? "0" : "initial",
        width: this.width || "300px",
        "max-width": this.maxWidth || "75vw",
        "--sidebar-transform-component-internal":
          this.side === "right" ? "100%" : "-100%",
        "--translate-x-component-internal": `${
          this.side === "right" ? "" : "-"
        }${this.touchMoveX}%`
      };
    }
  },
  methods: {
    close(e) {
      this.visibleLocal = null;
      const overlay = this.$refs["overlay"];
      if (overlay) {
        overlay.style["pointer-events"] = "none";
        const doc = document.elementFromPoint(e.clientX, e.clientY);
        if (doc.click) doc.click();
      }
    },
    touchstart(e) {
      this.touchStartX = e.changedTouches[0].clientX;
    },
    touchend(e) {
      if (this.$refs.sidebar) {
        if (this.touchMoveX > 25) {
          this.$refs.sidebar.style.transition = "";
          this.visibleLocal = null;
        } else if (this.touchMoveX == 0) {
          this.$refs.sidebar.style.transition = "";
        } else {
          this.$refs.sidebar.style.transition = "transform .2s";
          this.touchMoveX = 0;
        }
      }
    },
    touchmove(e) {
      const move = e.changedTouches[0].clientX;
      const width = window.screen.width;
      const delta = ((this.touchStartX - move) * 100) / width;
      if (this.side === "right") {
        this.touchMoveX = delta < 0 ? -delta : 0;
      } else {
        this.touchMoveX = delta;
      }
    }
  }
};
</script> 