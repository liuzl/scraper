($ => {
    "use strict";

    window.ext = function (opts) {

        let loadingInfo = {};
        let isReloading = false;

        /*
         * ################################
         * PUBLIC
         * ################################
         */
        this.initialized = null;
        this.firstRun = true;
        this.refreshRun = true;
        this.isDev = false;
        this.elements = {};
        this.opts = opts;

        /**
         * Constructor
         */
        this.run = () => {
            this.isDev = this.opts.manifest.version_name === "Dev" || !('update_url' in this.opts.manifest);
            initHelpers();

            this.helper.model.init().then(() => {
                return this.helper.i18n.init();
            }).then(() => {
                destroyOldInstance();

                this.helper.font.init();
                this.helper.stylesheet.init();
                this.helper.stylesheet.addStylesheets(["content"]);

                return initSidebar();
            }).then(() => {
                if (this.elements.iframe && this.elements.iframe[0]) { // prevent errors on pages which instantly redirect and prevent the iframe from loading this way
                    this.helper.toggle.init();
                    this.helper.list.init();
                    this.helper.scroll.init();
                    this.helper.tooltip.init();
                    this.helper.sidebarEvents.init();
                    this.helper.dragndrop.init();
                    this.helper.keyboard.init();

                    if (document.referrer === "") {
                        this.helper.model.call("addViewAmount", {url: location.href});
                    }
                }
            });
        };

        /**
         * Reloads the sidebar, the model data, the language variables and the bookmark list
         */
        this.reload = () => {
            if (isReloading === false) { // prevent multiple reload attempts -> only proceed if the previous run finished
                isReloading = true;
                this.helper.model.init().then(() => {
                    return Promise.all([
                        this.helper.i18n.init(),
                        this.helper.entry.init()
                    ]);
                }).then(() => {
                    let data = this.helper.model.getData(["u/entriesLocked", "u/showHidden"]);

                    if (data.entriesLocked === false) {
                        this.elements.sidebar.addClass(opts.classes.sidebar.entriesUnlocked);
                    } else {
                        this.elements.sidebar.removeClass(opts.classes.sidebar.entriesUnlocked);
                    }

                    this.helper.list.updateBookmarkBox().then(() => {
                        this.helper.search.init();
                    });
                });
            }
        };

        /**
         * Tracks some events of the initial state of the extension,
         * is called after opening the sidebar and is only executed the first time the sidebar is opened
         */
        this.trackInitialEvents = () => {
            let trackEvents = () => {
                let sort = this.helper.list.getSort();
                this.helper.model.call("trackEvent", {
                    category: "sorting",
                    action: "initial",
                    label: sort.name + "_" + sort.dir
                });

                let searchVal = this.elements.header.find("div." + opts.classes.sidebar.searchBox + " > input[type='text']")[0].value;
                if (searchVal.length > 0) {
                    this.helper.model.call("trackEvent", {
                        category: "search",
                        action: "search",
                        label: "initial",
                        value: searchVal.length
                    });
                }
            };

            if (this.firstRun) { // extension is not loaded yet -> wait for the loaded event (happens when sidebar is automatically opened on new tab page)
                $(document).on(opts.events.loaded + ".bs", () => {
                    trackEvents();
                });
            } else { // extension is loaded -> track events
                trackEvents();
            }
        };

        /**
         * Prints the given parameters in the console (only if this.dev = true)
         *
         * @param {mixed} msg
         */
        this.log = (...msg) => {
            if (this.isDev) {
                console.log(...msg);
            }
        };

        /**
         * Sets a class to the iframe body and fires an event to indicate, that the extension is loaded completely
         */
        this.loaded = () => {
            if (!this.elements.iframeBody.hasClass(opts.classes.sidebar.extLoaded)) {
                let data = this.helper.model.getData(["b/pxTolerance", "a/showIndicator"]);
                this.elements.iframeBody.addClass(opts.classes.sidebar.extLoaded);
                this.helper.list.updateSidebarHeader();
                this.helper.search.init();
                this.initialized = +new Date();

                this.log(this.initialized - this.updateBookmarkBoxStart);

                this.helper.utility.triggerEvent("loaded", {
                    pxTolerance: data.pxTolerance,
                    showIndicator: data.showIndicator
                });
            }

            isReloading = false;
        };

        /**
         * Adds a loading mask over the sidebar
         */
        this.startLoading = () => {
            this.elements.sidebar.addClass(opts.classes.sidebar.loading);

            if (loadingInfo.timeout) {
                clearTimeout(loadingInfo.timeout);
            }
            if (typeof loadingInfo.loader === "undefined" || loadingInfo.loader.length() === 0) {
                loadingInfo.loader = this.helper.template.loading().appendTo(this.elements.sidebar);
            }
        };

        /**
         * Removes the loading mask after the given time
         *
         * @param {int} timeout in ms
         */
        this.endLoading = (timeout = 500) => {
            loadingInfo.timeout = setTimeout(() => {
                this.elements.sidebar.removeClass(opts.classes.sidebar.loading);
                loadingInfo.loader && loadingInfo.loader.remove();
                loadingInfo = {};
            }, timeout);
        };

        /**
         * Initialises the not yet loaded images in the sidebar
         */
        this.initImages = () => {
            if (this.elements.iframe.hasClass(opts.classes.page.visible)) {
                this.elements.sidebar.find("img[" + opts.attr.src + "]").forEach((_self) => {
                    let img = $(_self);
                    let src = img.attr(opts.attr.src);
                    img.removeAttr(opts.attr.src);
                    img.attr("src", src);
                });
            }
        };

        /**
         * Adds a mask over the sidebar to notice that the page needs to be reloaded to make the sidebar work again
         */
        this.addReloadMask = () => {
            this.elements.sidebar.text("");
            let reloadMask = $("<div />").attr("id", opts.ids.sidebar.reloadInfo).prependTo(this.elements.sidebar);
            let contentBox = $("<div />").prependTo(reloadMask);

            $("<p />").html(this.helper.i18n.get("status_background_disconnected_reload_desc")).appendTo(contentBox);
            $("<a />").text(this.helper.i18n.get("status_background_disconnected_reload_action")).appendTo(contentBox);
        };

        /**
         * Adds a mask over the sidebar to encourage the user the share their userdata
         */
        this.addShareUserdataMask = () => {
            this.elements.sidebar.find("#" + opts.ids.sidebar.shareUserdata).remove();
            let shareUserdataMask = $("<div />").attr("id", opts.ids.sidebar.shareUserdata).prependTo(this.elements.sidebar);
            let contentBox = $("<div />").prependTo(shareUserdataMask);

            $("<h2 />").html(this.helper.i18n.get("share_userdata_headline")).appendTo(contentBox);
            $("<p />").html(this.helper.i18n.get("share_userdata_desc")).appendTo(contentBox);
            $("<p />").html(this.helper.i18n.get("share_userdata_desc2")).appendTo(contentBox);
            $("<p />").addClass(opts.classes.sidebar.shareUserdataNotice).html(this.helper.i18n.get("share_userdata_notice")).appendTo(contentBox);

            $("<a />").data("accept", true).text(this.helper.i18n.get("share_userdata_accept")).appendTo(contentBox);
            $("<a />").data("accept", false).text(this.helper.i18n.get("share_userdata_decline")).appendTo(contentBox);
        };


        /*
         * ################################
         * PRIVATE
         * ################################
         */

        /**
         * Initialises the helper objects
         */
        let initHelpers = () => {
            this.helper = {
                model: new window.ModelHelper(this),
                toggle: new window.ToggleHelper(this),
                entry: new window.EntryHelper(this),
                list: new window.ListHelper(this),
                scroll: new window.ScrollHelper(this),
                template: new window.TemplateHelper(this),
                i18n: new window.I18nHelper(this),
                font: new window.FontHelper(this),
                sidebarEvents: new window.SidebarEventsHelper(this),
                search: new window.SearchHelper(this),
                stylesheet: new window.StylesheetHelper(this),
                dragndrop: new window.DragDropHelper(this),
                checkbox: new window.CheckboxHelper(this),
                keyboard: new window.KeyboardHelper(this),
                overlay: new window.OverlayHelper(this),
                utility: new window.UtilityHelper(this),
                specialEntry: new window.SpecialEntryHelper(this),
                contextmenu: new window.ContextmenuHelper(this),
                tooltip: new window.TooltipHelper(this)
            };
        };

        /**
         * Removes the existing instance of the extension
         */
        let destroyOldInstance = () => {
            let sidebarIframe = $("iframe#" + opts.ids.page.iframe);

            if (sidebarIframe.length() > 0) {
                this.log("DESTROY");

                sidebarIframe.remove();
                $("iframe#" + opts.ids.page.overlay).remove();
                $("div#" + opts.ids.page.indicator).remove();

                $(document).off("*.bs");
                $(window).off("*.bs");
            }
        };

        /**
         * Creates the basic html markup for the sidebar and the visual
         *
         * @returns {Promise}
         */
        let initSidebar = async () => {
            this.elements.iframe = $('<iframe id="' + opts.ids.page.iframe + '" />').appendTo("body");

            if (this.helper.model.getData("b/animations") === false) {
                this.elements.iframe.addClass(opts.classes.page.noAnimations);
            }

            this.elements.iframeBody = this.elements.iframe.find("body");
            this.elements.sidebar = $('<section id="' + opts.ids.sidebar.sidebar + '" />').appendTo(this.elements.iframeBody);
            this.elements.bookmarkBox = {};

            ["all", "search"].forEach((val) => {
                this.elements.bookmarkBox[val] = this.helper.scroll.add(opts.ids.sidebar.bookmarkBox[val], $("<ul />").appendTo(this.elements.sidebar));
            });

            this.elements.filterBox = $("<div />").addClass(opts.classes.sidebar.filterBox).appendTo(this.elements.sidebar);

            this.elements.iframeBody.attr(opts.attr.dragCancel, this.helper.i18n.get("sidebar_drag_cancel"));
            this.elements.header = $("<header />").prependTo(this.elements.sidebar);
            this.helper.stylesheet.addStylesheets(["sidebar"], this.elements.iframe);

            let data = this.helper.model.getData(["u/entriesLocked", "u/showHidden", "a/darkMode"]);

            if (data.entriesLocked === false) {
                this.elements.sidebar.addClass(opts.classes.sidebar.entriesUnlocked);
            }

            if (data.darkMode === true) {
                this.elements.iframeBody.addClass(opts.classes.page.darkMode);
            }
        };
    };

})(jsu);