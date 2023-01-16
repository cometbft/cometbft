<template>
  <div>
    <modal @visible="sidebarShow" v-bind="{visible, side: 'center', boxShadow: '0 10px 25px 0 rgba(0,0,0,.15)'}">
      <div class="container">
        <div class="contents">
          <div class="contents__icon">
            <div class="icon">
              <svg class="icon__img" xmlns="http://www.w3.org/2000/svg" width="100%" height="100%" viewBox="0 0 24 24" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path><line x1="12" y1="9" x2="12" y2="13"></line><line x1="12" y1="17" x2="12.01" y2="17"></line></svg>
            </div>
          </div>
          <div>
            <div class="h1">Deactivate account</div>
            <div class="p">Are you sure you want to deactivate your account? All of your data will be permanantly removed. This action cannot be undone.</div>
          </div>
        </div>
        <div class="footer">
          <div class="button__list">
            <btn background="regular" size="s" @click.native="visible = false" @keydown.enter="visible = false">Cancel</btn>
            <btn background="danger" size="s" @click.native="visible = false" @keydown.enter="visible = false">Deactivate</btn>
          </div>
        </div>
      </div>
    </modal>
  </div>
</template>

<style scoped>
.container {
  font-family: var(--ds-font-family, inherit);
}
.contents {
  display: grid;
  grid-template-columns: 40px 1fr;
  gap: 1rem;
  padding: 1.5rem;
}
.icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: rgb(224, 36, 36, .1);
  stroke: #e02424;
  fill: none;
}
.icon__img {
  width: 50%;
  transform: translateY(-1px);
}
.h1 {
  font-weight: 400;
  font-size: 1.15rem;
  margin-bottom: .75rem;
  color: rgba(0,0,0,.875);
}
.p {
  margin-top: .75rem;
  font-size: .875rem;
  line-height: 1.5;
  color: rgba(0,0,0,.5);
}
.footer {
  background: rgba(0,0,0,.05);
  padding: 1rem 1.5rem;
}
.button__list {
  display: grid;
  justify-content: flex-end;
  grid-auto-flow: column;
  gap: 1rem;
  font-size: .875rem;
}

@media screen and (max-width: 700px) {
  .contents {
    grid-template-columns: 1fr;
    text-align: center;
  }
  .contents__icon {
    display: flex;
    justify-content: center;
  }
  .button__list {
    grid-auto-flow: row;
    justify-content: stretch;
    font-size: 1rem;
    gap: .75rem;
  }
}
</style>

<script>
import Modal from "./Modal.vue"
import Button from "../Button/Button.vue"

export default {
  components: { Modal, btn: Button },
  data: function() {
    return {
      visible: true
    }
  },
  methods: {
    sidebarShow() {
      setTimeout(() => {
        this.visible = true
      }, 500)
    },
  },
}
</script>