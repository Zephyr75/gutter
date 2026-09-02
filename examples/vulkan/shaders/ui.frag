#version 450
#extension GL_EXT_nonuniform_qualifier : require

layout(location = 0) in vec2 inUV;
layout(location = 1) in vec4 inColor;
layout(location = 2) flat in int inTex;

layout(location = 0) out vec4 outColor;

// Bindless combined-image-sampler array, one slot per Texture.Key the host has
// seen. Partially bound: slots the host has not filled are never indexed.
layout(set = 0, binding = 0) uniform sampler2D textures[];

void main() {
    vec4 c = inColor;
    if (inTex >= 0) {
        // Both sides are straight alpha, so tinting is a plain multiply.
        c *= texture(textures[nonuniformEXT(inTex)], inUV);
    }
    outColor = c;
}
