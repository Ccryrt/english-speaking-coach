(() => {
  const select=document.querySelector('#practice-language');
  if(!select)return;
  const current=location.pathname.startsWith('/ja/')?'ja':'en';
  select.value=current;
  select.addEventListener('change',async()=>{
    const chosen=select.value,error=document.querySelector('#language-error');
    select.disabled=true;error.hidden=true;
    try {
      const response=await fetch('/api/language',{cache:'no-store',signal:AbortSignal.timeout(10000)});
      const state=await response.json();
      if(!response.ok)throw new Error(state.error||'无法读取语言设置');
      const saved=await fetch('/api/language',{method:'POST',headers:{'Content-Type':'application/json','X-Coach-Token':state.token},body:JSON.stringify({language:chosen}),signal:AbortSignal.timeout(90000)});
      const result=await saved.json();
      if(!saved.ok)throw new Error(result.error||'切换失败');
      // Voice/run/session IDs belong to one archive: do not carry them into another language.
      const page=location.hash.slice(1).split('?')[0];
      const safe=['overview','live','sessions','terms','stats','storage'].includes(page)?page:'overview';
      location.assign('/'+chosen+'/#'+safe);
    } catch(e) { select.value=current;error.textContent='未切换：'+e.message;error.hidden=false;select.disabled=false; }
  });
})();
