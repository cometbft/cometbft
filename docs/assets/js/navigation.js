window.addEventListener("DOMContentLoaded", initNavigationTree);
window.addEventListener("scroll", handlePageScroll);

function initNavigationTree() {
  document
    .querySelectorAll(".PrimaryNavigation-button")
    .forEach((buttonElement) => {
      buttonElement.addEventListener("click", clickPrimaryNavigationButton);
    });
}

function clickPrimaryNavigationButton(event) {
  const buttonElement = event.target;
  const parentMenuItem = buttonElement.closest(".PrimaryNavigation-item");

  event.preventDefault();

  parentMenuItem.classList.toggle("is-active");
}

function updateTableOfContents() {
  const allHeadings = Array.from(
    document.querySelectorAll("h2[id], h3[id], h4[id], h5[id], h6[id]")
  );

  let lowestSectionAboveFold = null;

  allHeadings.reverse().every((heading) => {
    const headingId = heading.id;
    const coords = heading.getBoundingClientRect();

    if (coords.top <= 100) {
      lowestSectionAboveFold = headingId;
      return false;
    }

    return true;
  });

  const allAnchors = Array.from(
    document.querySelectorAll(".TableOfContents a")
  );

  allAnchors.forEach((anchor) => {
    anchor.classList.toggle(
      "is-active",
      anchor.href.endsWith(`#${lowestSectionAboveFold}`)
    );
  });
}

function updateStickyHeader(scrollEvent) {
  const siteHeaderElement = document.querySelector(".SiteHeader");
  const distanceScrolled = scrollEvent.target.scrollingElement.scrollTop;

  siteHeaderElement.classList.toggle("is-stuck", distanceScrolled >= 50);
}

function handlePageScroll(scrollEvent) {
  updateTableOfContents(scrollEvent);
  updateStickyHeader(scrollEvent);
}
