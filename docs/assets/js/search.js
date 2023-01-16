window.addEventListener("DOMContentLoaded", initSiteSearch);

function initSiteSearch() {
  window.appState = {
    refs: {
      inputElement: document.querySelector(".js-search-input"),
      searchResultsModal: document.querySelector(".js-search-results"),
      timer: null,
    },
    pages: [],
    results: [],
  };

  window.appState.refs.inputElement.addEventListener(
    "focus",
    handleFirstInputFocus
  );
}

function handleFirstInputFocus(focusEvent) {
  const { inputElement } = window.appState.refs;

  inputElement.addEventListener("keyup", (keyUpEvent) => {
    clearTimeout(window.appState.refs.timer);

    window.appState.refs.timer = setTimeout(() => {
      handleKeyUp(keyUpEvent);
    }, 300);
  });

  inputElement.removeEventListener("focus", handleFirstInputFocus);

  inputElement.addEventListener("focus", handleInputFocus);

  handleInputFocus(focusEvent);

  loadSearchData();
}

function handleInputFocus(focusEvent) {}

function handleKeyUp(keyUpEvent) {
  render();
}

function sanitizeSearchData(page) {
  return {
    ...page,
    title: decodeURIComponent(page.title).replace(/\+/g, " "),
    content: decodeURIComponent(page.content).replace(/\+/g, " "),
  };
}

function loadSearchData() {
  fetch("/assets/js/search.data.json")
    .then((response) => response.json())
    .then((pages) => {
      window.appState.pages = pages.map(sanitizeSearchData);
    });
}

function render() {
  const { inputElement, searchResultsModal } = window.appState.refs;

  const keywords = inputElement.value.toLowerCase().split(" ");

  const matchingPages = [];

  searchResultsModal.innerHTML = ``;

  if (!keywords.length) {
    return;
  }

  window.appState.pages.forEach((page) => {
    const lowerCaseTitle = page.title.toLowerCase();

    let matchScore = 0;
    let numOccurrencesInContent = 0;

    keywords.forEach((keyword) => {
      if (lowerCaseTitle.includes(keyword)) {
        matchScore += 10;
      }

      if (page.content.includes(keyword)) {
        matchScore += 1;
        numOccurrencesInContent = (
          page.content.match(new RegExp(keyword, "g")) || []
        ).length;
      }
    });

    if (matchScore > 0) {
      matchingPages.push({
        ...page,
        numOccurrencesInContent,
        score: matchScore,
      });
    }
  });

  matchingPages
    .sort((a, b) => a.numOccurrencesInContent - b.numOccurrencesInContent)
    .sort((a, b) => a.score - b.score)
    .reverse();

  const renderedResults = matchingPages
    .map(
      (result) => `
        <li class="SiteSearch-result js-search-result">
          <a class="SiteSearch-resultButton" href="${result.url}">
            <span class="SiteSearch-resultTitle">${result.title}</span>
            <span class="SiteSearch-resultDescription">${result.url}</span>
          </a>
        </li>
      `
    )
    .join("");

  searchResultsModal.innerHTML = matchingPages.length
    ? `
      <li class="SiteSearch-resultTally">
        ${matchingPages.length} ${
        matchingPages.length === 1 ? `Page` : `Pages`
      } Matched
      </li>
      ${renderedResults}
    `
    : `
      <li class="SiteSearch-result--empty">
        No Results
      </li>
    `;
}
