import PageHeader from "./PageHeader.vue";

export default {
  title: "PageHeader",
  component: PageHeader
};

export const normal = () => ({
  components: { PageHeader },
  template: `
    <div>
      <page-header>
        <div slot="title">Title</div>
        <div slot="subtitle">This is an example of subtitle.</div>
      </page-header>
    </div>
  `
});

export const withProjectTitle = () => ({
  components: { PageHeader },
  template: `
    <div>
      <page-header>
        <div slot="project-title">Project Title</div>
        <div slot="title">This is an example of title.</div>
      </page-header>
    </div>
  `
});

export const withProjectSuptitle = () => ({
  components: { PageHeader },
  template: `
    <div>
      <page-header>
        <div slot="suptitle">Coming soon &#x2026;</div>
        <div slot="title">Title</div>
        <div slot="subtitle">Subtitle</div>
      </page-header>
    </div>
  `
});