/*! (c) Philipp König under GPL-3.0 */
(e=>{"use strict";window.TemplateHelper=function(t){this.loading=(()=>e('<svg class="loading" width="36px" height="36px" viewBox="0 0 36 36" xmlns="http://www.w3.org/2000/svg"><circle fill="none" stroke-width="3" stroke-linecap="round" cx="18" cy="18" r="16"></circle></svg>')),this.svgByName=(t=>new Promise((s,n)=>{e.xhr(chrome.extension.getURL("img/"+t+".svg")).then(e=>{s(e.responseText)},()=>{n()})})),this.footer=(()=>{let t=e('<footer> <a id="copyright" href="https://extensions.blockbyte.de/" target="_blank">  &copy; <span class="created">2016</span>&ensp;<strong>Blockbyte</strong> </a></footer>'),s=+t.find("span.created").text(),n=(new Date).getFullYear();return n>s&&t.find("span.created").text(s+" - "+n),t})}})(jsu);