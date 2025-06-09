(()=>{function p(l,o=300){let a;return(...s)=>{clearTimeout(a),a=setTimeout(()=>{l.apply(this,s)},o)}}var i={attachFiles:"Attach Files",bold:"Bold",bullets:"Bullets",byte:"Byte",bytes:"Bytes",captionPlaceholder:"Add a caption\u2026",code:"Code",heading1:"Heading",indent:"Increase Level",italic:"Italic",link:"Link",numbers:"Numbers",outdent:"Decrease Level",quote:"Quote",redo:"Redo",remove:"Remove",strike:"Strikethrough",undo:"Undo",unlink:"Unlink",url:"URL",urlPlaceholder:"Enter a URL\u2026",GB:"GB",KB:"KB",MB:"MB",PB:"PB",TB:"TB"};Trix.config.toolbar.getDefaultHTML=()=>`<div class="trix-button-row">
      <span class="trix-button-group trix-button-group--text-tools" data-trix-button-group="text-tools">
        <button type="button" class="trix-button trix-button--icon trix-button--icon-bold" data-trix-attribute="bold" data-trix-key="b" title="${i.bold}" tabindex="-1">${i.bold}</button>
        <button type="button" class="trix-button trix-button--icon trix-button--icon-italic" data-trix-attribute="italic" data-trix-key="i" title="${i.italic}" tabindex="-1">${i.italic}</button>
        <button type="button" class="trix-button trix-button--icon trix-button--icon-strike" data-trix-attribute="strike" title="${i.strike}" tabindex="-1">${i.strike}</button>
        <button type="button" class="trix-button trix-button--icon trix-button--icon-link" data-trix-attribute="href" data-trix-action="link" data-trix-key="k" title="${i.link}" tabindex="-1">${i.link}</button>
      </span>

      <span class="trix-button-group trix-button-group--block-tools" data-trix-button-group="block-tools">
        <button type="button" class="trix-button trix-button--icon trix-button--icon-quote" data-trix-attribute="quote" title="${i.quote}" tabindex="-1">${i.quote}</button>
        <button type="button" class="trix-button trix-button--icon trix-button--icon-code" data-trix-attribute="code" title="${i.code}" tabindex="-1">${i.code}</button>
        <button type="button" class="trix-button trix-button--icon trix-button--icon-bullet-list" data-trix-attribute="bullet" title="${i.bullets}" tabindex="-1">${i.bullets}</button>
        <button type="button" class="trix-button trix-button--icon trix-button--icon-number-list" data-trix-attribute="number" title="${i.numbers}" tabindex="-1">${i.numbers}</button>
        <button type="button" class="trix-button trix-button--icon trix-button--icon-decrease-nesting-level" data-trix-action="decreaseNestingLevel" title="${i.outdent}" tabindex="-1">${i.outdent}</button>
        <button type="button" class="trix-button trix-button--icon trix-button--icon-increase-nesting-level" data-trix-action="increaseNestingLevel" title="${i.indent}" tabindex="-1">${i.indent}</button>
      </span>

      <span class="trix-button-group-spacer"></span>

      <span class="trix-button-group trix-button-group--history-tools" data-trix-button-group="history-tools">
        <button type="button" class="trix-button trix-button--icon trix-button--icon-undo" data-trix-action="undo" data-trix-key="z" title="${i.undo}" tabindex="-1">${i.undo}</button>
        <button type="button" class="trix-button trix-button--icon trix-button--icon-redo" data-trix-action="redo" data-trix-key="shift+z" title="${i.redo}" tabindex="-1">${i.redo}</button>
      </span>
    </div>

    <div class="trix-dialogs" data-trix-dialogs>
      <div class="trix-dialog trix-dialog--link" data-trix-dialog="href" data-trix-dialog-attribute="href">
        <div class="trix-dialog__link-fields">
          <input type="url" name="href" class="trix-input trix-input--dialog" placeholder="${i.urlPlaceholder}" aria-label="${i.url}" data-trix-validate-href required data-trix-input>
          <div class="trix-button-group">
            <input type="button" class="trix-button trix-button--dialog" value="${i.link}" data-trix-method="setAttribute">
            <input type="button" class="trix-button trix-button--dialog" value="${i.unlink}" data-trix-method="removeAttribute">
          </div>
        </div>
      </div>
    </div>`;(function(){document.addEventListener("htmx:beforeRequest",t=>{let e=t.target.getAttribute("data-window");if(e){let n=document.getElementById(e);n&&(b(t),n.remove(),t.preventDefault())}}),document.addEventListener("htmx:afterProcessNode",t=>{if(t.target.classList.contains("shifu-window")||t.target.classList.contains("shifu-window-overlay")){let e=t.target;e.classList.contains("shifu-window-overlay")&&(e=e.children[0]),document.querySelectorAll(".shifu-window").forEach(r=>r.classList.remove("shifu-window-active")),e.classList.add("shifu-window-active");let n=e.querySelector(".shifu-window-title");if(n){let r=localStorage.getItem(e.id);if(r){let x=JSON.parse(r);e.style.top=x.top,e.style.left=x.left}n.addEventListener("mousedown",d),n.addEventListener("mouseup",c)}}}),document.addEventListener("htmx:beforeCleanupElement",b),document.addEventListener("htmx:trigger",t=>{if(t.target.classList.contains("shifu-window-close")){b(t);let e=t.target;for(;!e.classList.contains("shifu-window");)e=e.parentNode;e.parentNode&&e.parentNode.classList.contains("shifu-window-overlay")&&e.parentNode.remove(),e.remove()}}),window.addEventListener("mousemove",f);let l=p(t=>{localStorage.setItem(t.id,JSON.stringify({top:t.style.top,left:t.style.left}))}),o=null,a,s,u=0;function d(t){document.querySelectorAll(".shifu-window").forEach(e=>e.classList.remove("shifu-window-active")),o=t.target.parentNode,o.style.zIndex>u&&(u=o.style.zIndex),u++,o.style.zIndex=u,o.classList.add("shifu-window-active"),a=t.clientX,s=t.clientY}function c(){o=null}function f(t){if(o){let e=a-t.clientX,n=s-t.clientY;a=t.clientX,s=t.clientY,o.style.top=o.offsetTop-n+"px",o.style.left=o.offsetLeft-e+"px",l(o)}}function b(t){if(t.target.classList&&t.target.classList.contains("shifu-window")){let e=t.target.querySelector(".shifu-window-title");e&&(e.removeEventListener("mousedown",d),e.removeEventListener("mouseup",c))}}document.addEventListener("htmx:afterRequest",t=>{let e=t.target.getAttribute("data-window");if(e){let n=document.querySelector(e);if(!n)return;for(;!n.classList.contains("shifu-window");)n=n.parentNode;let r=n.querySelector(".shifu-window-title");r&&(r.removeEventListener("mousedown",d),r.removeEventListener("mouseup",c)),n.parentNode&&n.parentNode.classList.contains("shifu-window-overlay")&&n.parentNode.remove(),n.remove()}})})();})();
//# sourceMappingURL=admin.js.map
