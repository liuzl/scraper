($ => {
    "use strict";

    window.SearchHelper = function (ext) {

        let searchTimeout = null;

        /**
         * Initializes the helper
         */
        this.init = () => {
            let searchValue = ext.helper.model.getData("u/searchValue");

            if (searchValue && searchValue.length > 0) { // search value is not empty -> restore
                ext.elements.header.addClass(ext.opts.classes.sidebar.searchVisible);
                handleSearchValChanged(searchValue);
            }

            initEvents();
        };

        /**
         * Clears the search field and shows the normal bookmark list again
         *
         * @returns {Promise}
         */
        this.clearSearch = () => {
            return new Promise((resolve) => {
                ext.helper.contextmenu.close();
                ext.helper.tooltip.close();
                ext.elements.header.removeClass(ext.opts.classes.sidebar.searchVisible);
                handleSearchValChanged("").then(resolve);
            });
        };

        /**
         * Handles the view of the search result list
         *
         * @param val
         * @returns {Promise}
         */
        let handleSearchValChanged = (val = null) => {
            return new Promise((resolve) => {
                let searchField = ext.elements.header.find("div." + ext.opts.classes.sidebar.searchBox + " > input[type='text']");
                if (val === null) {
                    val = searchField[0].value;
                } else {
                    searchField[0].value = val;
                }

                if (val && val.length > 0) { // search field is not empty
                    handleSearch(searchField, val).then(resolve);
                } else { // empty search field -> reset list
                    handleSearchReset(searchField).then(resolve);
                }
            });
        };

        /**
         * Updates the search results if search value is not the same as currently visible
         *
         * @param {jsu} searchField
         * @param {string} val
         * @returns {Promise}
         */
        let handleSearch = (searchField, val) => {
            return new Promise((resolve) => {
                let isFirstRun = ext.firstRun;
                ext.elements.bookmarkBox["all"].removeClass(ext.opts.classes.sidebar.active).removeClass(ext.opts.classes.scrollBox.scrolled);
                ext.elements.bookmarkBox["search"].addClass(ext.opts.classes.sidebar.active);
                ext.helper.scroll.focus();
                ext.helper.list.updateSortFilter();

                if (val !== searchField.data("lastVal")) { // search value is not the same
                    ext.startLoading();
                    searchField.data("lastVal", val);

                    if (searchTimeout) {
                        clearTimeout(searchTimeout);
                        searchTimeout = null;
                    }

                    ext.helper.entry.initOnce().then(() => {
                        return ext.helper.model.setData({"u/searchValue": val});
                    }).then(() => {
                        ext.helper.scroll.setScrollPos(ext.elements.bookmarkBox["search"], 0);
                        return ext.helper.model.call("searchBookmarks", {searchVal: val});
                    }).then((response) => {
                        ext.elements.bookmarkBox["search"].children("p").remove();

                        let hasResults = false;
                        let list = ext.elements.bookmarkBox["search"].children("ul");
                        list.text("");

                        if (response.bookmarks && response.bookmarks.length > 0) { // results for your search value
                            hasResults = ext.helper.list.addBookmarkDir(response.bookmarks, list, false);
                        }

                        if (hasResults === false) { // no results
                            $("<p />").text(ext.helper.i18n.get("sidebar_search_no_results")).appendTo(ext.elements.bookmarkBox["search"]);
                        }

                        if (!isFirstRun) {
                            searchTimeout = setTimeout(() => {
                                ext.helper.model.call("trackEvent", {
                                    category: "search",
                                    action: "search",
                                    label: "search",
                                    value: val.length
                                });
                            }, 500);
                        }

                        ext.endLoading(500);
                        resolve();
                    });
                }
            });
        };

        /**
         * Resets the search results and shows the normal bookmark list again
         *
         * @param {jsu} searchField
         * @returns {Promise}
         */
        let handleSearchReset = (searchField) => {
            return new Promise((resolve) => {
                searchField.removeData("lastVal");

                ext.helper.model.setData({"u/searchValue": false}).then(() => {
                    if (ext.elements.bookmarkBox["search"].hasClass(ext.opts.classes.sidebar.active)) {
                        ext.startLoading();
                        ext.elements.bookmarkBox["all"].addClass(ext.opts.classes.sidebar.active);
                        ext.elements.bookmarkBox["search"].removeClass(ext.opts.classes.sidebar.active);
                        ext.helper.scroll.restoreScrollPos(ext.elements.bookmarkBox["all"]);
                        ext.helper.scroll.focus();
                        ext.endLoading();
                    }

                    ext.helper.list.updateSortFilter();
                    resolve();
                });
            });
        };


        /**
         * Initializes the events for the search field
         */
        let initEvents = () => {
            ext.elements.header.on("click", "a." + ext.opts.classes.sidebar.search, (e) => {
                e.preventDefault();
                e.stopPropagation();
                ext.helper.contextmenu.close();
                ext.helper.tooltip.close();
                ext.elements.header.addClass(ext.opts.classes.sidebar.searchVisible);
                ext.elements.header.find("div." + ext.opts.classes.sidebar.searchBox + " > input[type='text']")[0].focus();
            });

            ext.elements.header.on("keyup", "div." + ext.opts.classes.sidebar.searchBox + " > input[type='text']", (e) => {
                e.preventDefault();
                handleSearchValChanged();
            });

            ext.elements.header.on("click", "a." + ext.opts.classes.sidebar.searchClose, (e) => {
                e.preventDefault();
                e.stopPropagation();
                this.clearSearch();
            });
        };
    };

})(jsu);