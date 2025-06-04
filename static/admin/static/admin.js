(()=>{function p(s,n=300){let r;return(...a)=>{clearTimeout(r),r=setTimeout(()=>{s.apply(this,a)},n)}}var i={attachFiles:"Attach Files",bold:"Bold",bullets:"Bullets",byte:"Byte",bytes:"Bytes",captionPlaceholder:"Add a caption\u2026",code:"Code",heading1:"Heading",indent:"Increase Level",italic:"Italic",link:"Link",numbers:"Numbers",outdent:"Decrease Level",quote:"Quote",redo:"Redo",remove:"Remove",strike:"Strikethrough",undo:"Undo",unlink:"Unlink",url:"URL",urlPlaceholder:"Enter a URL\u2026",GB:"GB",KB:"KB",MB:"MB",PB:"PB",TB:"TB"};Trix.config.toolbar.getDefaultHTML=()=>`<div class="trix-button-row">
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
    </div>`;(function(){document.addEventListener("htmx:beforeRequest",e=>{let t=e.target.getAttribute("data-window");if(t){let o=document.getElementById(t);o&&(c(e),o.remove(),e.preventDefault())}}),document.addEventListener("htmx:afterProcessNode",e=>{if(e.target.classList.contains("shifu-window")||e.target.classList.contains("shifu-window-overlay")){let t=e.target;t.classList.contains("shifu-window-overlay")&&(t=t.children[0]);let o=t.querySelector(".shifu-window-title");if(o){let b=localStorage.getItem(t.id);if(b){let x=JSON.parse(b);t.style.top=x.top,t.style.left=x.left}o.addEventListener("mousedown",l),o.addEventListener("mouseup",d)}}}),document.addEventListener("htmx:beforeCleanupElement",c),document.addEventListener("htmx:trigger",e=>{if(e.target.classList.contains("shifu-window-close")){c(e);let t=e.target;for(;!t.classList.contains("shifu-window");)t=t.parentNode;t.parentNode&&t.parentNode.classList.contains("shifu-window-overlay")&&t.parentNode.remove(),t.remove()}}),window.addEventListener("mousemove",f);let s=p(e=>{localStorage.setItem(e.id,JSON.stringify({top:e.style.top,left:e.style.left}))}),n=null,r,a,u=0;function l(e){n=e.target.parentNode,n.style.zIndex>u&&(u=n.style.zIndex),u++,n.style.zIndex=u,r=e.clientX,a=e.clientY}function d(){n=null}function f(e){if(n){let t=r-e.clientX,o=a-e.clientY;r=e.clientX,a=e.clientY,n.style.top=n.offsetTop-o+"px",n.style.left=n.offsetLeft-t+"px",s(n)}}function c(e){if(e.target.classList&&e.target.classList.contains("shifu-window")){let t=e.target.querySelector(".shifu-window-title");t&&(t.removeEventListener("mousedown",l),t.removeEventListener("mouseup",d))}}window.shifuCloseWindow=e=>{let t=document.querySelector(e);if(!t)return;for(;!t.classList.contains("shifu-window");)t=t.parentNode;let o=t.querySelector(".shifu-window-title");o&&(o.removeEventListener("mousedown",l),o.removeEventListener("mouseup",d)),t.parentNode&&t.parentNode.classList.contains("shifu-window-overlay")&&t.parentNode.remove(),t.remove()}})();})();
//# sourceMappingURL=admin.js.map
