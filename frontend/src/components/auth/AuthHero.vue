<template>
  <section class="auth-hero">
    <p class="eyebrow">WsAI 智能工作台</p>
    <h1>先完成认证，再进入你的对话控制台。</h1>
    <p class="hero-copy">
      当前后端支持邮箱验证码注册、用户名密码登录，以及基于 JWT 的受保护聊天与图片识别接口。
    </p>

    <div class="hero-cards">
      <article>
        <span>01</span>
        <strong>邮箱注册</strong>
        <p>邮箱收到验证码后即可注册，成功后会直接返回登录令牌。</p>
      </article>
      <article>
        <span>02</span>
        <strong>账号登录</strong>
        <p>使用邮箱里收到的用户名和密码登录，前端会自动保存令牌。</p>
      </article>
      <article>
        <span>03</span>
        <strong>进入对话</strong>
        <p>可查看历史会话、流式聊天，并在工具区直接做图片识别。</p>
      </article>
    </div>
  </section>
</template>
