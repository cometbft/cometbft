<template>
  <div>
    <div class="shortcuts">
      <div class="shortcuts__h1">Keyboard shortcuts</div>
      <div class="shortcuts__table">
        <div class="shortcuts__table__row">
          <div class="shortcuts__table__row__keys">
            <div class="shortcuts__table__row__keys__item">/</div>
          </div>
          <div class="shortcuts__table__row__desc">Open search window</div>
        </div>
        <div class="shortcuts__table__row">
          <div class="shortcuts__table__row__keys">
            <div class="shortcuts__table__row__keys__item" style="font-size: .65rem;">esc</div>
          </div>
          <div class="shortcuts__table__row__desc">Close search window</div>
        </div>
        <div class="shortcuts__table__row">
          <div class="shortcuts__table__row__keys">
            <div class="shortcuts__table__row__keys__item">↵</div>
          </div>
          <div class="shortcuts__table__row__desc">Open highlighted search result</div>
        </div>
        <div class="shortcuts__table__row">
          <div class="shortcuts__table__row__keys">
            <div class="shortcuts__table__row__keys__item" style="font-size: .65rem;">▼</div>
            <div class="shortcuts__table__row__keys__item" style="font-size: .65rem;">▲</div>
          </div>
          <div class="shortcuts__table__row__desc">Navigate between search results</div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.shortcuts {
  display: flex;
  justify-content: center;
  flex-direction: column;
  padding-top: 4rem;
  padding-bottom: 4rem;
  font-family: var(--ds-font-family, inherit);
}
.shortcuts__h1 {
  text-align: center;
  color: rgba(22,25,49,0.65);
  text-transform: uppercase;
  letter-spacing: 0.2em;
  font-size: 0.75rem;
  margin-bottom: 2rem;
}
.shortcuts__table__row {
  display: grid;
  grid-template-columns: 3fr 7fr;
  gap: 1.5rem;
  align-items: center;
  margin-top: 0.5rem;
  margin-bottom: 0.5rem;
}
.shortcuts__table__row__keys {
  display: flex;
  justify-content: flex-end;
}
.shortcuts__table__row__keys__item {
  color: #46509f;
  background-color: rgba(176,180,207,0.2);
  border: 1px solid rgba(176,180,207,0.09);
  border-radius: 0.25rem;
  font-size: 0.8125rem;
  width: 1.5rem;
  height: 1.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 2px;
}
.shortcuts__table__row__desc {
  color: rgba(22,25,49,0.65);
}
</style>