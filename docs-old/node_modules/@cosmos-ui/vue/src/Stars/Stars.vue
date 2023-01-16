<template>
  <div>
    <div class="container">
      <canvas ref="canvas" :width="canvas.width" :height="canvas.height"></canvas>
    </div>
  </div>
</template>

<style scoped>
.container {
  background: radial-gradient(#18154c, #08091e 80%);
  height: 500px;;
}
</style>

<script>
import { isEqual } from "lodash"

export default {
  props: {
    vertical: {
      type: Number
    },
    horizontal: {
      type: Number
    },
    flickering: {
      type: Number
    }
  },
  data: function() {
    return {
      canvas: {
        width: null,
        height: null,
      },
      context: null,
      requestAnimationFrameId: null,
      canvasResize: null
    }
  },
  mounted() {
    const circlePath = new Path2D("M5 2.5C5 3.88071 3.88071 5 2.5 5C1.11929 5 0 3.88071 0 2.5C0 1.11929 1.11929 0 2.5 0C3.88071 0 5 1.11929 5 2.5Z")
    let timestampLast = 0;
    const scrollHorizontally = (timestamp, c) => {
      if (timestamp - timestampLast > 10) {
        for (let i=0; i < stars.length; i++) {
          const
            opacity = Math.sin(timestamp * stars[i].random / (this.flickering || 600)),
            xPosNew = stars[i].x > this.canvas.width ? 0 : stars[i].x + (this.horizontal || .25),
            yPosNew = stars[i].y +  Math.sin(opacity / (this.vertical || 20)) * (stars[i].random > .5 ? 1 : -1)
          stars[i] = {
            ...stars[i],
            x: xPosNew,
            y: yPosNew
          }
        }
        timestampLast = timestamp
      }
      c.clearRect(0, 0, this.canvas.width, this.canvas.height)
      for (let i = 0; i < stars.length; i++) {
        c.fillStyle = `rgba(${[...stars[i].color, Math.sin(timestamp * stars[i].random / (this.flickering || 600))].join(",")})`
        c.translate(stars[i].x , stars[i].y)
        c.fill(circlePath)
        c.translate(-stars[i].x , -stars[i].y)
      }
    }
    const renderStart = (timestamp) => {
      let c = this.context
      scrollHorizontally(timestamp, c)
      this.requestAnimationFrameId = requestAnimationFrame(renderStart)
    }
    const canvasResize = () => {
      stars = starsGenerate()
      this.canvas.width = this.canvasCalc().width
      this.canvas.height = this.canvasCalc().height
    }
    const starsGenerate = () => {
      const colors = [[227, 109, 37], [205, 6, 10], [208, 222, 204], [67, 189, 155], [220, 189, 45] ]
      let result = []
      for (let i=0; i < this.canvasCalc().width / 3; i++) {
        const
          x = Math.floor(Math.random() * this.canvasCalc().width),
          y =  Math.floor(Math.random() * this.canvasCalc().height)
        result.push({
          x,
          y,
          color: colors[Math.floor(Math.random() * colors.length)],
          random: Math.random()
        })
      }
      return result
    }
    canvasResize()
    this.canvasResize = canvasResize
    let stars = starsGenerate();
    const canvas = this.$refs.canvas
    this.context = canvas.getContext("2d")
    window.addEventListener("resize", canvasResize, false)
    renderStart()
  },
  beforeDestroy() {
    window.removeEventListener("resize", this.canvasResize, false)
    window.cancelAnimationFrame(this.requestAnimationFrameId)
  },
  methods: {
    canvasCalc() {
      const el = this.$refs.canvas.parentNode
      const style = getComputedStyle(el);
      const height = el.clientHeight - parseFloat(style.paddingTop) - parseFloat(style.paddingBottom);
      const width = el.clientWidth - parseFloat(style.paddingLeft) - parseFloat(style.paddingRight);
      return {height, width}
    },
  }
}
</script>