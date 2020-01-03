# Fast Mesh Segmentation

Meant to take large files (multiple GBs worth / 100million+ tris) and segment them into multiple smaller files for sake of loading them into game engines.

WIP

## Features

Why is this is better than using other programs:

* Only loads what FBX nodes are needed for mesh segmentation. Ignores all other fbx data, saving on RAM. 
* Combines all geometries found in the FBX, always exports only 2 model files (or one if the clipping plane collides with nothing)

## Results

```txt
2020/01/02 22:26:40 Loading Model: dragon_vrip.fbx took 524.9901ms
2020/01/02 22:26:41 Splitting model by plane took 564.9936ms
2020/01/02 22:26:43 Saving Model: retained.obj took 2.6189968s
2020/01/02 22:26:49 Saving Model: clipped.obj took 5.4730392s
```

![Results](https://i.imgur.com/QCW2qzq.png)

## Credits

### OSS 

Credits to [o5h](https://github.com/o5h/fbx/tree/3a77542940a3e1fb404bfd00f2e49565a504a2df) and [three.js](https://github.com/mrdoob/three.js/blob/de530d6bae1bf40d1e001411bc3e02a915c2c993/examples/js/loaders/FBXLoader.js) for helping me get started with code. Credits to blender for a detailed [explanation to the fbx format](https://code.blender.org/2013/08/fbx-binary-file-format-specification/).

### Other Resources

* [Clipping a Mesh Against a Plane](https://www.geometrictools.com/Documentation/ClipMesh.pdf)