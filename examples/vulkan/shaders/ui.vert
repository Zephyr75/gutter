#version 450

// One quad per gutter Cmd. There is no vertex buffer: the six corners are
// generated from gl_VertexIndex and placed by the push constant, so a Cmd
// becomes a Draw and nothing else.
layout(push_constant) uniform Push {
    vec4 color;   //  0  straight-alpha tint (the fill colour when tex < 0)
    vec2 screen;  // 16  framebuffer size in pixels
    vec2 pos;     // 24  Cmd.Rect top-left, pixels
    vec2 size;    // 32  Cmd.Rect size, pixels
    int  tex;     // 40  bindless slot, or -1 for a solid rect
} pc;

layout(location = 0) out vec2 outUV;
layout(location = 1) out vec4 outColor;
layout(location = 2) flat out int outTex;

const vec2 corners[6] = vec2[6](
    vec2(0, 0), vec2(1, 0), vec2(1, 1),
    vec2(0, 0), vec2(1, 1), vec2(0, 1));

void main() {
    vec2 uv = corners[gl_VertexIndex];
    vec2 px = pc.pos + uv * pc.size;
    // gutter's origin is top-left with Y down, which is also Vulkan's NDC
    // orientation, so this is a plain scale and bias -- no Y flip.
    gl_Position = vec4(px / pc.screen * 2.0 - 1.0, 0.0, 1.0);
    outUV = uv;
    outColor = pc.color;
    outTex = pc.tex;
}
